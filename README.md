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
    
    

    

