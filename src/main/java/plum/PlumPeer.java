package plum;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import java.io.IOException;
import java.util.logging.Logger;
import java.util.ArrayList;
import java.net.InetAddress;
import java.net.UnknownHostException;

public class PlumPeer {
	private static final Logger logger = Logger.getLogger(PlumPeer.class.getName());
	private final Server server;
	private final int port;
	private final String host;
	private ArrayList<IPAddress> addressBook;
	private ArrayList<Transaction> memPool;

	// initializer suite
	public PlumPeer() {
		this("localhost", 50051);
	}

	public PlumPeer(int port) {
		this("localhost", port);
	}

	public PlumPeer(String host, int port) {
		this.host = host;
		this.port = port;
		addressBook = new ArrayList<IPAddress>();
		memPool = new ArrayList<Transaction>();
		server = ServerBuilder.forPort(port).addService(new PlumServiceImpl(this)).build();
	}

	private void start() throws IOException {
		server.start();
		logger.info("Server started, listening on " + port);
		notifyToConductor();
		Runtime.getRuntime().addShutdownHook(new Thread(() -> {
			System.err.println("*** shutting down gRPC server since JVM is shutting down");
			PlumPeer.this.stop();
			System.err.println("*** server shut down");
		}));
	}

	// server related features
	private void stop() {
		if (server != null)
			server.shutdown();
	}

	private void blockUntilShutdown() throws InterruptedException {
		if (server != null) {
			server.awaitTermination();
		}
	}

	// notify this peer's participation to conductor
	private void notifyToConductor() {
		// current conductor's address
		PlumClient toConductor = new PlumClient("localhost", 50055);

		// add my address to conductor
		try {
			InetAddress local = InetAddress.getLocalHost();
			String myIp = local.getHostAddress().toString();
			toConductor.addAddress(myIp, 50051);
			Thread.sleep(3000);
		} catch (UnknownHostException e) {
			e.printStackTrace();
		} catch (InterruptedException e) {
			e.printStackTrace();
		}

		// get all addressbook from conductor
		this.addressBook = toConductor.getAddressBook();
	}

	// getters and setters
	public int getPort() {
		return this.port;
	}

	public String getHost() {
		return this.host;
	}

	public ArrayList<IPAddress> getAddressBook() {
		return this.addressBook;
	}

	public ArrayList<Transaction> getMemPool() {
		return this.memPool;
	}

	// entry point of PlumPeer
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
