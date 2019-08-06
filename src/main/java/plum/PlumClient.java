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
	private final PlumServiceGrpc.PlumServiceBlockingStub blockingStub;

	public PlumClient(String host, int port) {
		this(ManagedChannelBuilder.forAddress(host, port).usePlaintext().build());
	}

	PlumClient(ManagedChannel channel) {
		this.channel = channel;
		blockingStub = PlumServiceGrpc.newBlockingStub(channel);
	}

	public void shutdown() throws InterruptedException {
		channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
	}

	/*
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
	*/

	public void getIP() {
		logger.info("Will try to get IP from gRPC server");
		Empty req = Empty.newBuilder().build();
		IPAddress res;
		try {
			res = blockingStub.getIP(req);
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
			return;
		}
		logger.info("IP: " + res.getAddress());
	}

	public static void main(String[] args) throws Exception {
		PlumClient client = new PlumClient("localhost", 50051);
		try {
			client.getIP();
		} finally {
			client.shutdown();
		}
	}
}
