# go-peekerconn

## Usecases

[Pinecone I2P](https://github.com/BieHDC/dendrite/blob/main/cmd/dendrite-demo-pinecone-i2p/main.go#L415)

## Why?

Sometimes you find yourself in the situation where you want to see the data inside a connection and take action based on it. In my concrete example it is ether a HTTP request or a Pinecone Handshake. Depending on the type i pass it other over to a HTTP Server or down the Pinecone router to establish a connection.

## Ok, but where is the issue?

Once you have read from a connection the bytes are consumed and they are not returning. Peekerconn works around that by making an io.Multireader from the consumed bytes and the rest of a net.Conn. Then the path down the line that consumes the packet will again read the whole packet as if nothing happened.

## Is there another solution?

Maybe, and i challange you to do it better. Let the benchmark wars begin.