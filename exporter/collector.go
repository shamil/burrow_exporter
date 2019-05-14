package exporter

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

// If we are missing a status, it will return 0
var Status = map[string]int{
	"NOTFOUND": 1,
	"OK":       2,
	"WARN":     3,
	"ERR":      4,
	"STOP":     5,
	"STALL":    6,
	"REWIND":   7,
}

var (
	kafkaConsumerPartitionLagDesc           = prometheus.NewDesc("kafka_burrow_partition_lag", "The lag of the latest offset commit on a partition as reported by burrow.", []string{"cluster", "group", "topic", "partition"}, nil)
	kafkaConsumerPartitionCurrentOffsetDesc = prometheus.NewDesc("kafka_burrow_partition_current_offset", "The latest offset commit on a partition as reported by burrow.", []string{"cluster", "group", "topic", "partition"}, nil)
	kafkaConsumerPartitionCurrentStatusDesc = prometheus.NewDesc("kafka_burrow_partition_status", "The status of a partition as reported by burrow.", []string{"cluster", "group", "topic", "partition"}, nil)
	kafkaConsumerPartitionMaxOffsetDesc     = prometheus.NewDesc("kafka_burrow_partition_max_offset", "The log end offset on a partition as reported by burrow.", []string{"cluster", "group", "topic", "partition"}, nil)
	kafkaConsumerTotalLagDesc               = prometheus.NewDesc("kafka_burrow_total_lag", "The total amount of lag for the consumer group as reported by burrow.", []string{"cluster", "group"}, nil)
	kafkaConsumerStatusDesc                 = prometheus.NewDesc("kafka_burrow_status", "The status of a partition as reported by burrow.", []string{"cluster", "group"}, nil)
	kafkaTopicPartitionOffsetDesc           = prometheus.NewDesc("kafka_burrow_topic_partition_offset", "The latest offset on a topic's partition as reported by burrow.", []string{"cluster", "topic", "partition"}, nil)
)

type Collector struct {
	client *BurrowClient
	mutex  sync.Mutex

	skipPartitionStatus        bool
	skipConsumerStatus         bool
	skipPartitionLag           bool
	skipPartitionCurrentOffset bool
	skipPartitionMaxOffset     bool
	skipTotalLag               bool
	skipTopicPartitionOffset   bool
}

func (c *Collector) processGroup(cluster, group string) (metrics []prometheus.Metric) {
	resp, err := c.client.ConsumerGroupLag(cluster, group)
	if err != nil {
		log.With("err", err).Errorf("Error getting lag for consumer group (%v)", group)
		return
	}

	commonLabels := []string{resp.Status.Cluster, resp.Status.Group}

	for _, partition := range resp.Status.Partitions {

		labels := append(commonLabels, partition.Topic, strconv.Itoa(int(partition.Partition)))

		if !c.skipPartitionLag {
			metric, err := prometheus.NewConstMetric(
				kafkaConsumerPartitionLagDesc,
				prometheus.GaugeValue,
				float64(partition.CurrentLag),
				labels...,
			)

			if err != nil {
				log.With("err", err).Errorf("Failed to create metric")
			} else {
				metrics = append(metrics, metric)
			}
		}

		if !c.skipPartitionCurrentOffset {
			metric, err := prometheus.NewConstMetric(
				kafkaConsumerPartitionCurrentOffsetDesc,
				prometheus.GaugeValue,
				float64(partition.End.Offset),
				labels...,
			)

			if err != nil {
				log.With("err", err).Errorf("Failed to create metric")
			} else {
				metrics = append(metrics, metric)
			}
		}

		if !c.skipPartitionStatus {
			metric, err := prometheus.NewConstMetric(
				kafkaConsumerPartitionCurrentStatusDesc,
				prometheus.GaugeValue,
				float64(Status[partition.Status]),
				labels...,
			)

			if err != nil {
				log.With("err", err).Errorf("Failed to create metric")
			} else {
				metrics = append(metrics, metric)
			}
		}

		if !c.skipPartitionMaxOffset {
			metric, err := prometheus.NewConstMetric(
				kafkaConsumerPartitionMaxOffsetDesc,
				prometheus.GaugeValue,
				float64(partition.End.MaxOffset),
				labels...,
			)

			if err != nil {
				log.With("err", err).Errorf("Failed to create metric")
			} else {
				metrics = append(metrics, metric)
			}
		}
	}

	if !c.skipTotalLag {
		metric, err := prometheus.NewConstMetric(
			kafkaConsumerTotalLagDesc,
			prometheus.GaugeValue,
			float64(resp.Status.TotalLag),
			commonLabels...,
		)

		if err != nil {
			log.With("err", err).Errorf("Failed to create metric")
		} else {
			metrics = append(metrics, metric)
		}
	}

	if !c.skipConsumerStatus {
		metric, err := prometheus.NewConstMetric(
			kafkaConsumerStatusDesc,
			prometheus.GaugeValue,
			float64(Status[resp.Status.Status]),
			commonLabels...,
		)

		if err != nil {
			log.With("err", err).Errorf("Failed to create metric")
		} else {
			metrics = append(metrics, metric)
		}
	}

	return metrics
}

func (c *Collector) processTopic(cluster, topic string) (metrics []prometheus.Metric) {
	details, err := c.client.ClusterTopicDetails(cluster, topic)
	if err != nil {
		log.With("err", err).Errorf("Error getting details for cluster topic (%v)", topic)
		return
	}

	if !c.skipTopicPartitionOffset {
		for i, offset := range details.Offsets {
			labels := []string{cluster, topic, strconv.Itoa(i)}

			metric, err := prometheus.NewConstMetric(
				kafkaTopicPartitionOffsetDesc,
				prometheus.GaugeValue,
				float64(offset),
				labels...,
			)

			if err != nil {
				log.With("err", err).Errorf("Failed to create metric")
			} else {
				metrics = append(metrics, metric)
			}
		}
	}

	return metrics
}

func (c *Collector) scrape(cluster string) (metrics []prometheus.Metric) {
	groups, err := c.client.ListConsumers(cluster)
	if err != nil {
		log.With("err", err).Errorf("Error listing consumer groups (cluster: %v), skipping", cluster)
		groups = &ConsumerGroupsResp{}
	}

	for _, group := range groups.ConsumerGroups {
		metrics = append(metrics, c.processGroup(cluster, group)...)
	}

	topics, err := c.client.ListTopics(cluster)
	if err != nil {
		log.With("err", err).Errorf("Error listing topics (cluster: %v), skipping", cluster)
		topics = &TopicsResp{}
	}

	for _, topic := range topics.Topics {
		metrics = append(metrics, c.processTopic(cluster, topic)...)
	}

	return metrics
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	c.mutex.Lock()

	defer func() {
		c.mutex.Unlock()
		log.Infof("Finished scraping burrow, took %v.", time.Now().Sub(start))
	}()

	log.Info("Scraping burrow...")
	clusters, err := c.client.ListClusters()
	if err != nil {
		log.With("err", err).Error("Failed listing clusters")
		return
	}

	for _, cluster := range clusters.Clusters {
		for _, metric := range c.scrape(cluster) {
			ch <- metric
		}
	}
}

func NewCollector(burrowUrl string, apiVersion int, disabledMetrics string) *Collector {
	disabledMetricsSet := make(map[string]bool)

	for _, v := range strings.Split(disabledMetrics, ",") {
		disabledMetricsSet[v] = true
	}

	return &Collector{
		client:                     NewBurrowClient(burrowUrl, apiVersion),
		skipPartitionStatus:        disabledMetricsSet["partition-status"],
		skipConsumerStatus:         disabledMetricsSet["consumer-status"],
		skipPartitionLag:           disabledMetricsSet["partition-lag"],
		skipPartitionCurrentOffset: disabledMetricsSet["partition-current-offset"],
		skipPartitionMaxOffset:     disabledMetricsSet["partition-max-offset"],
		skipTotalLag:               disabledMetricsSet["total-lag"],
		skipTopicPartitionOffset:   disabledMetricsSet["topic-partition-offset"],
	}
}
