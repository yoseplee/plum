package plum;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.logging.Logger;
import java.util.ArrayList;

public class PlumPeer {
	private static final Logger logger = Logger.getLogger(PlumPeer.class.getName());
	private ArrayList<String> addressBook;
	private Server server;
	private int port;
	private String host;

	public PlumPeer() {
		this("localhost", 50051);
	}

	public PlumPeer(int port) {
		this("localhost", port);
	}

	public PlumPeer(String host, int port) {
		this.host = host;
		this.port = port;
		addressBook = new ArrayList<String>();
		server = ServerBuilder.forPort(port).addService(new PlumServiceImpl(this)).build();
	}

	private void start() throws IOException {
		// int port = 50051;
		// server = ServerBuilder.forPort(port).addService(new PlumServiceImpl(this)).build().start();
		server.start();
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
}
