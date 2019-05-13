package exporter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"
)

type BurrowResp struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

type ClustersResp struct {
	BurrowResp
	Clusters []string `json:"clusters"`
}

type ClusterDetails struct {
	Brokers       []string `json:"brokers"`
	Zookeepers    []string `json:"zookeepers"`
	BrokerPort    int      `json:"broker_port"`
	ZookeeperPort int      `json:"zookeeper_port"`
	OffsetsTopic  string   `json:"offsets_topic"`
}

type ClusterDetailsResp struct {
	BurrowResp
	Cluster ClusterDetails `json:"cluster"`
}

type ConsumerGroupsResp struct {
	BurrowResp
	ConsumerGroups []string `json:"consumers"`
}

type TopicsResp struct {
	BurrowResp
	Topics []string `json:"topics"`
}

type ConsumerGroupTopicDetailsResp struct {
	BurrowResp
	Offsets []int64 `json:"offsets"`
}

type Offset struct {
	Offset    int64 `json:"offset"`
	Timestamp int64 `json:"timestamp"`
	Lag       int64 `json:"lag"`
	MaxOffset int64 `json:"max_offset"`
}

type ConsumerGroupStatus struct {
	Cluster    string      `json:"cluster"`
	Group      string      `json:"group"`
	Status     string      `json:"status"`
	MaxLag     Partition   `json:"maxlag"`
	Partitions []Partition `json:"partitions"`
	TotalLag   int64       `json:"totallag"`
}

type Partition struct {
	Topic      string `json:"topic"`
	Partition  int32  `json:"partition"`
	Status     string `json:"status"`
	Start      Offset `json:"start"`
	End        Offset `json:"end"`
	CurrentLag int64  `json:"current_lag"`
}

type ConsumerGroupStatusResp struct {
	BurrowResp
	Status ConsumerGroupStatus `json:"status"`
}

type ClusterTopicDetailsResp struct {
	BurrowResp
	Offsets []int64 `json:"offsets"`
}

type BurrowClient struct {
	baseURL    string
	apiversion int
	client     *http.Client
}

func (bc *BurrowClient) buildURL(endpoint string) (string, error) {
	parsedUrl, err := url.Parse(bc.baseURL)
	if err != nil {
		return "", err
	}

	parsedUrl.Path = path.Join(parsedUrl.Path, endpoint)

	return parsedUrl.String(), nil
}

func (bc *BurrowClient) getJsonReq(endpoint string, dest interface{}) error {
	resp, err := bc.client.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return err
	}

	return nil
}

func (bc *BurrowClient) HealthCheck() (bool, error) {
	endpoint, err := bc.buildURL("/burrow/admin")
	if err != nil {
		return false, err
	}

	if _, err := bc.client.Get(endpoint); err != nil {
		return false, err
	}

	return true, nil
}

func (bc *BurrowClient) ListClusters() (*ClustersResp, error) {
	endpoint, err := bc.buildURL("/kafka")
	if err != nil {
		return nil, err
	}

	clusters := &ClustersResp{}
	if err := bc.getJsonReq(endpoint, clusters); err != nil {
		return nil, err
	}

	if clusters.Error {
		return nil, fmt.Errorf(clusters.Message)
	}

	return clusters, nil
}

func (bc *BurrowClient) ClusterDetails(cluster string) (*ClusterDetailsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s", cluster))
	if err != nil {
		return nil, err
	}

	clusterDetails := &ClusterDetailsResp{}
	if err := bc.getJsonReq(endpoint, clusterDetails); err != nil {
		return nil, err
	}

	if clusterDetails.Error {
		return nil, fmt.Errorf(clusterDetails.Message)
	}

	return clusterDetails, nil
}

func (bc *BurrowClient) ListConsumers(cluster string) (*ConsumerGroupsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/consumer", cluster))
	if err != nil {
		return nil, err
	}

	consumers := &ConsumerGroupsResp{}
	if err := bc.getJsonReq(endpoint, consumers); err != nil {
		return nil, err
	}

	if consumers.Error {
		return nil, fmt.Errorf(consumers.Message)
	}

	return consumers, nil
}

func (bc *BurrowClient) ListConsumerTopics(cluster, consumerGroup string) (*TopicsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/consumer/%s/topic", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	consumerTopics := &TopicsResp{}
	if err := bc.getJsonReq(endpoint, consumerTopics); err != nil {
		return nil, err
	}

	if consumerTopics.Error {
		return nil, fmt.Errorf(consumerTopics.Message)
	}

	return consumerTopics, nil
}

func (bc *BurrowClient) ListTopics(cluster string) (*TopicsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/topic", cluster))
	if err != nil {
		return nil, err
	}

	consumerTopics := &TopicsResp{}
	if err := bc.getJsonReq(endpoint, consumerTopics); err != nil {
		return nil, err
	}

	if consumerTopics.Error {
		return nil, fmt.Errorf(consumerTopics.Message)
	}

	return consumerTopics, nil
}

func (bc *BurrowClient) ConsumerGroupTopicDetails(cluster, consumerGroup, topic string) (*ConsumerGroupTopicDetailsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/consumer/%s/topic/%s", cluster, consumerGroup, topic))
	if err != nil {
		return nil, err
	}

	topicDetails := &ConsumerGroupTopicDetailsResp{}
	if err := bc.getJsonReq(endpoint, topicDetails); err != nil {
		return nil, err
	}

	if topicDetails.Error {
		return nil, fmt.Errorf(topicDetails.Message)
	}

	return topicDetails, nil
}

func (bc *BurrowClient) ConsumerGroupStatus(cluster, consumerGroup string) (*ConsumerGroupStatusResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/consumer/%s/status", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	status := &ConsumerGroupStatusResp{}
	if err := bc.getJsonReq(endpoint, status); err != nil {
		return nil, err
	}

	if status.Error {
		return nil, fmt.Errorf(status.Message)
	}

	return status, nil
}

func (bc *BurrowClient) ConsumerGroupLag(cluster, consumerGroup string) (*ConsumerGroupStatusResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/consumer/%s/lag", cluster, consumerGroup))
	if err != nil {
		return nil, err
	}

	status := &ConsumerGroupStatusResp{}
	if err := bc.getJsonReq(endpoint, status); err != nil {
		return nil, err
	}

	if status.Error {
		return nil, fmt.Errorf(status.Message)
	}

	return status, nil
}

func (bc *BurrowClient) ClusterTopicDetails(cluster, topic string) (*ClusterTopicDetailsResp, error) {
	endpoint, err := bc.buildURL(fmt.Sprintf("/kafka/%s/topic/%s", cluster, topic))
	if err != nil {
		return nil, err
	}

	topicDetails := &ClusterTopicDetailsResp{}
	if err := bc.getJsonReq(endpoint, topicDetails); err != nil {
		return nil, err
	}

	if topicDetails.Error {
		return nil, fmt.Errorf(topicDetails.Message)
	}

	return topicDetails, nil
}

func NewBurrowClient(baseUrl string, apiVersion int) *BurrowClient {
	return &BurrowClient{
		baseURL:    fmt.Sprintf("%s/v%d", baseUrl, apiVersion),
		apiversion: apiVersion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
