---
layout: post
title:  "Go Lang com processamento assíncrono, utilizando Kafka"
categories: blog hard-skills go kafka
author: Carlos Lorenzon
comments: true
---
Sabemos que as soluções atualmente muitas vezes precisam utilizar processamento assíncrono, e existem várias ferramentas e linguagens para resolver isso.

Neste artigo, vou demonstrar de uma maneira muito simples como subiur um ambiente Kafka e usa-lo com a linguagem de programação Go.

## Primeiramente, o que é Kafka?

Kafka é uma solução de streaming de eventos distribuída de forma open-source. Ele fornece uma plataforma unificada, de alta taxa de transferência e baixa latência para lidar com fluxos de dados em tempo real. Kafka é usado para construir pipelines de dados em tempo real e aplicativos, sendo conhecido por sua durabilidade, escalabilidade e tolerância a falhas. Ele permite o desacoplamento de fluxos de dados e possibilita que os aplicativos publiquem, assinem, armazenem e processem fluxos de registros em tempo real.

## Agora que sabemos o que é Kafka, vamos preparar o ambiente para demonstrar o seu uso.

Para facilitar a visualização da produção e consumo de eventos em ação, vamos iniciar o serviço usando Docker Compose.

### Crie uma pasta chamada kafka-docker

```
mkdir kafka-docker
cd kafka-docker 
```

### Dentro desta pasta, crie o arquivo docker-compose.yml abaixo

```
version: '2'
services:

  broker:
    image: confluentinc/cp-kafka:7.6.1
    hostname: broker
    container_name: broker
    ports:
      - "9092:9092"
      - "9101:9101"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: 'CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT'
      KAFKA_ADVERTISED_LISTENERS: 'PLAINTEXT://broker:29092,PLAINTEXT_HOST://localhost:9092'
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS: 0
      KAFKA_TRANSACTION_STATE_LOG_MIN_ISR: 1
      KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR: 1
      KAFKA_JMX_PORT: 9101
      KAFKA_JMX_HOSTNAME: localhost
      KAFKA_PROCESS_ROLES: 'broker,controller'
      KAFKA_CONTROLLER_QUORUM_VOTERS: '1@broker:29093'
      KAFKA_LISTENERS: 'PLAINTEXT://broker:29092,CONTROLLER://broker:29093,PLAINTEXT_HOST://0.0.0.0:9092'
      KAFKA_INTER_BROKER_LISTENER_NAME: 'PLAINTEXT'
      KAFKA_CONTROLLER_LISTENER_NAMES: 'CONTROLLER'
      KAFKA_LOG_DIRS: '/tmp/kraft-combined-logs'
      # Replace CLUSTER_ID with a unique base64 UUID using "bin/kafka-storage.sh random-uuid" 
      # See https://docs.confluent.io/kafka/operations-tools/kafka-tools.html#kafka-storage-sh
      CLUSTER_ID: 'MkU3OEVBNTcwNTJENDM2Qk'

  schema-registry:
    image: confluentinc/cp-schema-registry:7.6.1
    hostname: schema-registry
    container_name: schema-registry
    depends_on:
      - broker
    ports:
      - "8081:8081"
    environment:
      SCHEMA_REGISTRY_HOST_NAME: schema-registry
      SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS: 'broker:29092'
      SCHEMA_REGISTRY_LISTENERS: http://0.0.0.0:8081

  connect:
    image: cnfldemos/cp-server-connect-datagen:0.6.4-7.6.0
    hostname: connect
    container_name: connect
    depends_on:
      - broker
      - schema-registry
    ports:
      - "8083:8083"
    environment:
      CONNECT_BOOTSTRAP_SERVERS: 'broker:29092'
      CONNECT_REST_ADVERTISED_HOST_NAME: connect
      CONNECT_GROUP_ID: compose-connect-group
      CONNECT_CONFIG_STORAGE_TOPIC: docker-connect-configs
      CONNECT_CONFIG_STORAGE_REPLICATION_FACTOR: 1
      CONNECT_OFFSET_FLUSH_INTERVAL_MS: 10000
      CONNECT_OFFSET_STORAGE_TOPIC: docker-connect-offsets
      CONNECT_OFFSET_STORAGE_REPLICATION_FACTOR: 1
      CONNECT_STATUS_STORAGE_TOPIC: docker-connect-status
      CONNECT_STATUS_STORAGE_REPLICATION_FACTOR: 1
      CONNECT_KEY_CONVERTER: org.apache.kafka.connect.storage.StringConverter
      CONNECT_VALUE_CONVERTER: io.confluent.connect.avro.AvroConverter
      CONNECT_VALUE_CONVERTER_SCHEMA_REGISTRY_URL: http://schema-registry:8081
      # CLASSPATH required due to CC-2422
      CLASSPATH: /usr/share/java/monitoring-interceptors/monitoring-interceptors-7.6.1.jar
      CONNECT_PRODUCER_INTERCEPTOR_CLASSES: "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
      CONNECT_CONSUMER_INTERCEPTOR_CLASSES: "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
      CONNECT_PLUGIN_PATH: "/usr/share/java,/usr/share/confluent-hub-components"
      CONNECT_LOG4J_LOGGERS: org.apache.zookeeper=ERROR,org.I0Itec.zkclient=ERROR,org.reflections=ERROR

  control-center:
    image: confluentinc/cp-enterprise-control-center:7.6.1
    hostname: control-center
    container_name: control-center
    depends_on:
      - broker
      - schema-registry
      - connect
      - ksqldb-server
    ports:
      - "9021:9021"
    environment:
      CONTROL_CENTER_BOOTSTRAP_SERVERS: 'broker:29092'
      CONTROL_CENTER_CONNECT_CONNECT-DEFAULT_CLUSTER: 'connect:8083'
      CONTROL_CENTER_CONNECT_HEALTHCHECK_ENDPOINT: '/connectors'
      CONTROL_CENTER_KSQL_KSQLDB1_URL: "http://ksqldb-server:8088"
      CONTROL_CENTER_KSQL_KSQLDB1_ADVERTISED_URL: "http://localhost:8088"
      CONTROL_CENTER_SCHEMA_REGISTRY_URL: "http://schema-registry:8081"
      CONTROL_CENTER_REPLICATION_FACTOR: 1
      CONTROL_CENTER_INTERNAL_TOPICS_PARTITIONS: 1
      CONTROL_CENTER_MONITORING_INTERCEPTOR_TOPIC_PARTITIONS: 1
      CONFLUENT_METRICS_TOPIC_REPLICATION: 1
      PORT: 9021

  ksqldb-server:
    image: confluentinc/cp-ksqldb-server:7.6.1
    hostname: ksqldb-server
    container_name: ksqldb-server
    depends_on:
      - broker
      - connect
    ports:
      - "8088:8088"
    environment:
      KSQL_CONFIG_DIR: "/etc/ksql"
      KSQL_BOOTSTRAP_SERVERS: "broker:29092"
      KSQL_HOST_NAME: ksqldb-server
      KSQL_LISTENERS: "http://0.0.0.0:8088"
      KSQL_CACHE_MAX_BYTES_BUFFERING: 0
      KSQL_KSQL_SCHEMA_REGISTRY_URL: "http://schema-registry:8081"
      KSQL_PRODUCER_INTERCEPTOR_CLASSES: "io.confluent.monitoring.clients.interceptor.MonitoringProducerInterceptor"
      KSQL_CONSUMER_INTERCEPTOR_CLASSES: "io.confluent.monitoring.clients.interceptor.MonitoringConsumerInterceptor"
      KSQL_KSQL_CONNECT_URL: "http://connect:8083"
      KSQL_KSQL_LOGGING_PROCESSING_TOPIC_REPLICATION_FACTOR: 1
      KSQL_KSQL_LOGGING_PROCESSING_TOPIC_AUTO_CREATE: 'true'
      KSQL_KSQL_LOGGING_PROCESSING_STREAM_AUTO_CREATE: 'true'

  ksqldb-cli:
    image: confluentinc/cp-ksqldb-cli:7.6.1
    container_name: ksqldb-cli
    depends_on:
      - broker
      - connect
      - ksqldb-server
    entrypoint: /bin/sh
    tty: true

  ksql-datagen:
    image: confluentinc/ksqldb-examples:7.6.1
    hostname: ksql-datagen
    container_name: ksql-datagen
    depends_on:
      - ksqldb-server
      - broker
      - schema-registry
      - connect
    command: "bash -c 'echo Waiting for Kafka to be ready... && \
                       cub kafka-ready -b broker:29092 1 40 && \
                       echo Waiting for Confluent Schema Registry to be ready... && \
                       cub sr-ready schema-registry 8081 40 && \
                       echo Waiting a few seconds for topic creation to finish... && \
                       sleep 11 && \
                       tail -f /dev/null'"
    environment:
      KSQL_CONFIG_DIR: "/etc/ksql"
      STREAMS_BOOTSTRAP_SERVERS: broker:29092
      STREAMS_SCHEMA_REGISTRY_HOST: schema-registry
      STREAMS_SCHEMA_REGISTRY_PORT: 8081

  rest-proxy:
    image: confluentinc/cp-kafka-rest:7.6.1
    depends_on:
      - broker
      - schema-registry
    ports:
      - 8082:8082
    hostname: rest-proxy
    container_name: rest-proxy
    environment:
      KAFKA_REST_HOST_NAME: rest-proxy
      KAFKA_REST_BOOTSTRAP_SERVERS: 'broker:29092'
      KAFKA_REST_LISTENERS: "http://0.0.0.0:8082"
      KAFKA_REST_SCHEMA_REGISTRY_URL: 'http://schema-registry:8081'
```

### Para rodar a solução, utilize o comando abaixo
```
docker-compose up -d
```

### Criando seu primeiro tópico

Agora que temos um ambiente Kafka em execução, vamos criar o seu primeiro tópico.

Para isso, acesse a URL: [http://localhost:9021/](http://localhost:9021/)

Neste paínel clique em: CONTROLCENTER.CLUSTER -> Topics -> Add topic


## Utilizando Kafka em Go Lang

Agora que temos nosso ambiente Kafka configurado, chegamos à parte divertida: vamos produzir e consumir eventos no Kafka.

### Produzindo eventos

Para produzir eventos em Kafka, utilizaremos a lib: [confluent kafka](https://github.com/confluentinc/confluent-kafka-go) da Confluent.

Para isso vamos começar nosso módulo Go, acesse a pasta que deseja deixar seu projeto e rode os comandos abaixo:


**Inicie o projeto**
```ssh
 go mod init github.com/seu_usuário/seurepo
```

**Adicione a lib sarama**
```
go get github.com/confluentinc/confluent-kafka-go/v2
```

Crie um arquivo chamado **kafka.go** e vamos implementar a produção e consumo dos eventos, para isso vamos criar a **function** main conforme abaixo:

```
package main

import (
	"fmt"
	"log"
	"os"	
)

const (
	topic         = "meu-topico"
	groupConsumer = "meu-grupo"
	broker        = "localhost:9092"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go [producer|consumer]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "producer":
		runProducer()
	case "consumer":
		runConsumer()
	default:
		fmt.Println("Modo não reconhecido. Use 'producer' ou 'consumer'")
	}

  func runProducer() {
  }

  func runConsumer() {
  }
}
```

Com esta implementação ao rodar o programa para testarmos, poderemos especificar qual funcionalidade queremos utilizar.

Vamos implementar a produção de eventos no trecho de código abaixo:
```
func runProducer() {
  config := &kafka.ConfigMap{
		"bootstrap.servers": broker,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	defer producer.Close()

	topicName := topic
	for _, word := range []string{"Hello", "Kafka", "World"} {
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}

		err := producer.Produce(message, nil)
		if err != nil {
			fmt.Printf("Failed to produce message: %s\n", err)
			continue
		}

		e := <-producer.Events()
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message to %v\n", m.TopicPartition)
		}
	}

	// Flush and wait for all messages to be delivered
	producer.Flush(15 * 1000)
}
```

Desta forma, se ocorrer como o esperado a mensagem de sucesso deve ser impressa no seu terminal ao executar o comando abaixo:

```
go run kafka.go producer
```

Vamos agora implementar o consumo dos eventos. Para simplificar este artigo, estamos criando um código considerando uma única partição (explicarei mais abaixo o que é uma partição).

```
func runConsumer() {
  topics := []string{topic}
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        broker,
		"broker.address.family":    "v4",
		"group.id":                 groupConsumer,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"enable.auto.offset.store": true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topics: %s\n", err)
		os.Exit(1)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}
			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}
```

Com esse código implementado, ao executar o comando abaixo é para imprimir o evento recebido.

```
go run kafka.go consumer
```

## Informação importante

### Offset

Lembre-se deste nome, pois ele é muito importante para o seu consumidor. No código criado acima, estamos passando o offset = `earliest`, o que isso significa? Significa que o seu consumidor vai pegar todos os eventos já publicados no tópico que ainda estejam dentro do TTL de vida dos eventos no mesmo.

Caso deseje começar a ler apenas os novos, pode usar `latest`, essa estratégia é útil para casos em que você precisa obter apenas os novos eventos a serem gerados, ignorando os publicados anteriormente à ativação do seu consumidor.

Se você não precisa garantir a ordem e deseja processar vários eventos simultaneamente, a próxima estratégia pode ser mais adequada.

### Controlando o último offset consumido pelo seu consumer

Para utilizar esta estratégia, você vai ter de controlar o commit dos seus eventos no `consumers groups` do Kafka

Ajustando a `config` do seu consumer para algo assim:

```
c, err := kafka.NewConsumer(&kafka.ConfigMap{
	"bootstrap.servers":        broker,
	"broker.address.family":    "v4",
	"group.id":                 groupConsumer,
	"session.timeout.ms":       6000,
	"auto.offset.reset":        "earliest",
	"enable.auto.commit":       false,
	"enable.auto.offset.store": false,
})
```

Analisando esta nova config, percebe-se que alteramos o `enable.auto.commit` e `enable.auto.offset.store` para que seja controlado pelo nosso consumer.

Além da alteração na config, necessitamos adicionar esse trecho no processamento do evento:
```
//Criar um TopicPartition com o offset específico
tp := kafka.TopicPartition{
  Topic:     e.TopicPartition.Topic,
  Partition: e.TopicPartition.Partition,
  Offset:    e.TopicPartition.Offset + 1, // Offset para ser commitado (próxima mensagem a ser lida)
}

// Commit manual do offset específico
offsets, err := c.CommitOffsets([]kafka.TopicPartition{tp})
if err != nil {
  fmt.Printf("Failed to commit offsets: %s\n", err)
} else {
  fmt.Printf("Committed offsets: %v\n", offsets)
}
```

Neste trecho acima, especificamos o offset a ser commitado, permitindo que o consumidor inicie o consumo de eventos a partir do último evento consumido anteriormente.

Com esta abordagem, podemos ajustar a estratégia de processamento de eventos usando Go Routines e apenas commitar o offset do último evento se não ocorrer nenhum erro durante o processamento.

É importante destacar que estamos consumindo eventos de uma única partição. Se desejar aumentar o paralelismo no processamento, é necessário ajustar o código para lidar com cada partição separadamente.

Caso não esteja familiarizado, aqui está uma breve explicação sobre o que é uma partição no Kafka.

### Partições
Uma partição em Kafka é uma subdivisão de um tópico, que permite a distribuição de dados e o paralelismo no processamento. Cada tópico pode ter várias partições, e cada partição é um log ordenado e imutável de eventos. As partições são distribuídas entre os brokers no cluster Kafka, permitindo escalabilidade e alta disponibilidade

Os eventos dentro de uma partição são atribuídos a offsets únicos, que identificam a posição do evento no log. Consumidores podem ler eventos de uma ou mais partições, e um grupo de consumidores pode se dividir entre as partições para processar os eventos em paralelo, aumentando a eficiência do processamento.


## Código final do artigo:
```
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	topic         = "meu-topico"
	groupConsumer = "meu-grupo"
	broker        = "localhost:9092"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go [producer|consumer]")
		return
	}

	mode := os.Args[1]

	switch mode {
	case "producer":
		runProducer()
	case "consumer":
		runConsumer()
	default:
		fmt.Println("Modo não reconhecido. Use 'producer' ou 'consumer'")
	}
}

func runConsumer() {
	topics := []string{topic}
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        broker,
		"broker.address.family":    "v4",
		"group.id":                 groupConsumer,
		"session.timeout.ms":       6000,
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Consumer %v\n", c)

	err = c.SubscribeTopics(topics, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to subscribe to topics: %s\n", err)
		os.Exit(1)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("%% Message on %s:\n%s\n",
					e.TopicPartition, string(e.Value))
				if e.Headers != nil {
					fmt.Printf("%% Headers: %v\n", e.Headers)
				}

				//Criar um TopicPartition com o offset específico
				tp := kafka.TopicPartition{
					Topic:     e.TopicPartition.Topic,
					Partition: e.TopicPartition.Partition,
					Offset:    e.TopicPartition.Offset + 1, // Offset para ser commitado (próxima mensagem a ser lida)
				}

				// Commit manual do offset específico
				offsets, err := c.CommitOffsets([]kafka.TopicPartition{tp})
				if err != nil {
					fmt.Printf("Failed to commit offsets: %s\n", err)
				} else {
					fmt.Printf("Committed offsets: %v\n", offsets)
				}

			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v: %v\n", e.Code(), e)
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}

	fmt.Printf("Closing consumer\n")
	c.Close()
}

func runProducer() {
	config := &kafka.ConfigMap{
		"bootstrap.servers": broker,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Failed to create producer: %s", err)
	}

	defer producer.Close()

	topicName := topic
	for _, word := range []string{"Hello", "Kafka", "World"} {
		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicName, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}

		err := producer.Produce(message, nil)
		if err != nil {
			fmt.Printf("Failed to produce message: %s\n", err)
			continue
		}

		e := <-producer.Events()
		m := e.(*kafka.Message)

		if m.TopicPartition.Error != nil {
			fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
		} else {
			fmt.Printf("Delivered message to %v\n", m.TopicPartition)
		}
	}

	// Flush and wait for all messages to be delivered
	producer.Flush(15 * 1000)
}
```

Espero que este artigo tenha sido útil.