package plum;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.net.InetAddress;
import java.util.logging.Logger;

public class PlumPeer {

	private static final Logger logger = Logger.getLogger(PlumPeer.class.getName());

	private Server server;

	private void start() throws IOException {
		int port = 50051;
		server = ServerBuilder.forPort(port).addService(new PlumServiceImpl()).build().start();
		logger.info("Server started, listening on " + port);
		Runtime.getRuntime().addShutdownHook(new Thread(() -> {
			System.err.println("*** shutting down gRPC server since JVM is shutting down");
			PlumPeer.this.stop();
			System.err.println("*** server shut down");
		}));
	}

	private void stop() {
		if (server != null)
			server.shutdown();
	}

	private void blockUntilShutdown() throws InterruptedException {
		if (server != null) {
			server.awaitTermination();
		}
	}

	public static void main(String[] args) {

		// run server on a thread. start and forgot
		Runnable serverTask = () -> {
			final PlumPeer server;
			server = new PlumPeer();
			try {
				server.start();
				server.blockUntilShutdown();
			} catch (IOException e) {
				e.printStackTrace();
			} catch (InterruptedException e) {
				e.printStackTrace();
			}
		};
		
		new Thread(serverTask).start();

	}

	static class PlumServiceImpl extends PlumServiceGrpc.PlumServiceImplBase {
		@Override
		public void sayHello(HelloRequest req, StreamObserver<HelloReply> responseObserver) {
			HelloReply reply = HelloReply.newBuilder().setMessage("Hello!").build();
			responseObserver.onNext(reply);
			responseObserver.onCompleted();
		}
	
		@Override
		public void getIP(Empty req, StreamObserver<IPAddress> responseObserver) {
			try {
				InetAddress local = InetAddress.getLocalHost();
				String myIp = local.getHostAddress().toString();
				IPAddress address = IPAddress.newBuilder().setAddress(myIp).build();
				responseObserver.onNext(address);
			} catch (Exception e) {
				System.err.println(e.getMessage());
			}
			responseObserver.onCompleted();
		}
	}
}
