package plum;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;

public class PlumClient {
	private static final Logger logger = Logger.getLogger(PlumClient.class.getName());
	private final ManagedChannel channel;
	private final TestGrpc.TestBlockingStub blockingStub;

	public PlumClient(String host, int port) {
		this(ManagedChannelBuilder.forAddress(host, port)
			.usePlaintext()
			.build());
	}

	PlumClient(ManagedChannel channel) {
		this.channel = channel;
		blockingStub = TestGrpc.newBlockingStub(channel);
	}

	public void shutdown() throws InterruptedException {
		channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
	}

	public void sayHello(String name) {
		logger.info("Will try to greet " + name + " ...");
		HelloRequest request = HelloRequest.newBuilder().setName(name).build();
		HelloReply response;
		try {
			response = blockingStub.sayHello(request);
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
      		return;
		}
		logger.info("Greeting: " + response.getMessage());
	}

	public static void main(String[] args) throws Exception {
		PlumClient client = new PlumClient("localhost", 50051);
		try {
			String user = "world";
			client.sayHello(user);
		} finally {
			client.shutdown();
		}
	}
}
