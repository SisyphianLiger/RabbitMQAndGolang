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

# Types of Exchanges
## Direct
Matches an exchange with a exact key
## Topic
Matches an exchange with a pattern match key
## Fanout
Routes messages to all of the queues bound to it ignoring routing keys
## Headers
Routes based on header values instead of  the routing key. Similar to topic but uses message header attributes for routing

# Routing Patterns
Most powerful feature of RabbitMQ
Allow message broker to flexibly tour messages to queues based on pattern matching rather than exact matches

## Words
Routing Keys -> made up of words separated by does: user.created or peril.game.won 

## Wildcards
    * --> substitutes for exactly one word
        peril.*.won

        MATCHES WITH
        peril.game.won
        peril.player.won

        NOT
        peril.game.lost
        peril.won

    # --> substitutes for zero or more words
        peril.# 
        MATCHS WITH
        peril.game.won
        peril.game.lost
        peril.player
        peril

        NOT
        peril.won
        perilgame.won
    
# Naming
You can name things whatever you want but its important to choose good names

## Exchange Naming
    - Its common for one system to use all the same exchange
        --> Thinkg how single database within postgres instance == single exchange 
## Queue Naming
    - key --> queue: If I have a routing key user.created, I might create a queue for my email notifier called:
        --> user.created.email_notifier
    - queue that consumes all events
        --> commen.created.<event_name>
    - Auto generated queue names are good for transient queues
## Routing Key Naming
    - Really want to get righ
    - key names NEED to be descriptive but flexible
    - noun.verb convention works well
# Dead Letter
PtP systems give an immediate message if something goes wrong: i.e. 404 Not Found
But what about RabbitMQ --> Asynchronous

## Dead Letter Exchanges and Queues
PubSub systems can aggregate messages that fail to be processed in a dead letter queue

Queues can be configured to send messages that faile to be processed to a dead letter exchange

# Ack and Nack
## Ack(nowledge)
When a consumer receives a message it must acknowlede it
If the subscriber crashes or fails to process the message the broker can just reque

### N(ot)ack(nowledged) --> Nack
Three options for a message
    1. Acknowledge: Processed Successfully
    2. Nack and requeue: Not processed successfully, but should be requeued on the same queue
    3. Nack and discard: nor processed successfully, and should be discarded to dead-letter queue


# Exact Delivery
What methods do we use to actually deliver messages (to the dead queue or just queues in general)
1. At-leat-once delivery: If message broker isn't sure the consumer receives msg, retry
    -> Complexity: Medium
    -> Efficiency: Medium
    -> Reliability: Medium
## At-least-once Cont.
    - This is the default
    - If consumer fails to proces message broker will just re-queue message
    - Think NackRequeue
2. At-most-one delivery: If message broker isn't sure the consumer revieced the message, discard
    -> Complexity: Low
    -> Efficiency: High
    -> Reliability: Low
## At-most-once Cont.
    - Messages that are not mission-critical
    - A debug log for example
3. Exactly-once deliver: message is guaranteed to be delivered once and only once
    -> Complexity: High
    -> Efficiency: Low  
    -> Reliability: High
## At-most-once Cont.
    - Apparenlty nearly impossible (distributed systems)
    - Also the slowest and most difficult to implement
# Serialization
Many Types:
    - JSON
    - Gob
    - Avro
    - buffers

Choosing one depends on schema of data used

## General rule of thumb
Adding/subtracting fields is ok I.E.:

```go
type User struct {
    ID int
    Name string
}
```

```go
type User struct {
    ID int
    Name string
    Email string
}
// or
type User struct {
    ID int
}
```

but changing the type of the field means we need to be careful

```go
type User struct {
    ID string // change to string
    Name string
}
```

This is because old messages with int ID's we push to our consumers will fail to be decoded

### Rule
If you make a breaking change to a schema, use a new routing-key/queue. That way the old consumers can polish off all the old messages and the new consumers can start fresh with a new schema

# Nodes and Clusters
In Production we have entire cluseter of nodes
    - High Availability
    - Scalability
    - Redundancy

Using A cluster of 3 nodes is a solid starting point for most production applications

Scale vertically when CPU/Ram/Disk hits

# Quorum Queues and Classic Queues

We have been using Classic Queues in this module, but they come with a problem

They are lightweight and pretty good in general, but if the server crashes...so does their queue.

To mitigate this we use Quorum Queues, a heavier alternative that is still able to maintain the queue if anything would go wrong.
