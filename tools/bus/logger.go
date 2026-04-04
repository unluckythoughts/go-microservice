package bus

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

func (b *bus) logMessage(source string, logEvent kafka.LogEvent) {
	l := b.l.Named(source)
	switch logEvent.Level {
	case 7, 6: // Debug
		l.Debugf("Kafka log: %s", logEvent.Message)
	case 5, 4: // Info
		l.Infof("Kafka log: %s", logEvent.Message)
	case 3, 2: // Warning
		l.Warnf("Kafka log: %s", logEvent.Message)
	case 1, 0: // Error
		l.Errorf("Kafka log: %s", logEvent.Message)
	default:
		// For unexpected levels, you might want to log them as well
		l.Infof("Kafka log (level %d): %s", logEvent.Level, logEvent.Message)
	}
}

func (b *bus) logKafkaMessages() {
	// consumer logs
	go func() {
		for logEvent := range b.kc.Logs() {
			b.logMessage("consumer", logEvent)
		}
	}()

	// producer logs
	go func() {
		for logEvent := range b.kp.Logs() {
			b.logMessage("producer", logEvent)
		}
	}()

}
