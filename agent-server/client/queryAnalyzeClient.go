package client

import (
	"context"
	"time"

	thrift "github.com/apache/thrift/lib/go/thrift"
	"agent-server/client/gen/query_server"
	Config "agent-server/config"
)

// QueryAnalyzeClient Query Server Thrift 客户端
type QueryAnalyzeClient struct {
	addr    string
	timeout time.Duration
}

// NewQueryAnalyzeClient 创建客户端（从配置读取）
func NewQueryAnalyzeClient() *QueryAnalyzeClient {
	cfg := Config.GetConfig()
	return &QueryAnalyzeClient{
		addr:    cfg.QueryAnalyze.Addr,
		timeout: time.Duration(cfg.QueryAnalyze.Timeout) * time.Second,
	}
}

// NewQueryAnalyzeClientWithAddr 创建客户端（指定地址）
func NewQueryAnalyzeClientWithAddr(addr string, timeout time.Duration) *QueryAnalyzeClient {
	return &QueryAnalyzeClient{
		addr:    addr,
		timeout: timeout,
	}
}

// connect 建立连接并返回客户端
func (c *QueryAnalyzeClient) connect() (thrift.TTransport, *query_server.QueryServerClient, error) {
	transport := thrift.NewTSocketConf(c.addr, &thrift.TConfiguration{
		ConnectTimeout: c.timeout,
		SocketTimeout:  c.timeout,
	})

	if err := transport.Open(); err != nil {
		return nil, nil, err
	}

	protocol := thrift.NewTBinaryProtocolConf(transport, nil)
	client := query_server.NewQueryServerClientProtocol(transport, protocol, protocol)

	return transport, client, nil
}

// Health 健康检查
func (c *QueryAnalyzeClient) Health() (string, error) {
	transport, client, err := c.connect()
	if err != nil {
		return "", err
	}
	defer transport.Close()

	return client.Health(context.Background())
}

// Segment 分词
func (c *QueryAnalyzeClient) Segment(texts []string, topK int32) ([]*query_server.SegmentResult_, error) {
	transport, client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer transport.Close()

	return client.Segment(context.Background(), texts, topK)
}

// NER 实体识别
func (c *QueryAnalyzeClient) NER(texts []string) ([]*query_server.NERResult_, error) {
	transport, client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer transport.Close()

	return client.Ner(context.Background(), texts)
}

// Embed 向量化
func (c *QueryAnalyzeClient) Embed(texts []string, normalize bool) ([][]float64, error) {
	transport, client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer transport.Close()

	return client.Embed(context.Background(), texts, normalize)
}

// Similarity 文本相似度
func (c *QueryAnalyzeClient) Similarity(text1, text2 string) (float64, error) {
	transport, client, err := c.connect()
	if err != nil {
		return 0, err
	}
	defer transport.Close()

	return client.Similarity(context.Background(), text1, text2)
}

// Analyze 综合分析
func (c *QueryAnalyzeClient) Analyze(texts []string, keywords, ner, embedding bool, topK int32) ([]*query_server.AnalysisResult_, error) {
	transport, client, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer transport.Close()

	return client.Analyze(context.Background(), texts, keywords, ner, embedding, topK)
}