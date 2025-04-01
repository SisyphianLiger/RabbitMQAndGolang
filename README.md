# RabbitMQ Module with "Game" --> Peril

I am currently learning on boot.dev about the ins and outs of RabbitMQ, and taking notes with terminology and key concepts I find interesting. Below are the sections of text and notes that can give some  more insight to my code base.

# What is Pub/Sub

An architecture that allows for three things

A Publisher (Sender)
An Intermediary (Message Broker)
A Subscriber(s) (Message Receiver)

# Point to Point
An Architecture that allows for communication from Client to Server

Bi-directional Communication
A Client (Sender/Receiver)
A Server (Sender/Receiver)

# AMQP Protocol
A communication protocol that defines how messages are sent, recieved, and understood. 

AMQP -> Advanced Message Queing Protocol is an open standard for passing business messages between applications and organizations

# MQTT
A Protocol for small IoT devices such as optimized to be lightweight

# STOMP
Designed for web applications and is designed to be simple and easy to use

# Exchanges and Queues

An Exchange is where a publisher sends messages typically with a routing key

Exchange takes message uses the routing key as a filter and sends the message to any queues that are listening for that routing key

Publisher --> Exchanges --> Queues

Publishers don't know about Queues

Terminology:
    - Exchange: A routing agent that sends messages to queues
    - Binding: A link between an exchange and a queue that uses a routing key to decided where messages go
    - Queue: A buffer in the RabbitMQ server that holds messages until they are consumed
    - Channel: A virtual connection inside a connection that allows you to create queues, exchanges, publish messages
    - Connectin: TCP connection to the RabbitMQ Server

# Queues
Can be durable or transient

Durable: stored on disk
Transient: stored in memory

# Transient Queues
Exclusive queue --> can only be used by the connection that created it
Auto-delete --> the queue will be automatically deleted when its last connection is closed



# Multi Consumers
A quese can have 0,1,many consumers
1. If queue has 0 --> messages will accumulate in the queue will never be processed
2. If queue has 1 --> consumer will process all messages in the queue
3. If queue has many --> messages will be distributed between them in round-robin fashion


