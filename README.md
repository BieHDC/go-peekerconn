# go-peekerconn

## Usecases

[Pinecone I2P](https://github.com/BieHDC/dendrite/blob/main/cmd/dendrite-demo-pinecone-i2p/main.go#L415)

## Why?

Sometimes you find yourself in the situation where you want to see the data inside a connection and take action based on it. In my concrete example it is ether a HTTP request or a Pinecone Handshake. Depending on the type i pass it other over to a HTTP Server or down the Pinecone router to establish a connection.