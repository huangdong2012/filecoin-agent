package kafka

type Options func(*Option)

type Option struct {
	Brokers []string
	Rest    bool
	Verbose bool
}

type OffsetOption struct {
	GetOffset func(topic string, partition int32) int64
	SetOffset func(topic string, partition int32, offset int64)
}
