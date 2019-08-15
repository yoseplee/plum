package plum;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import java.io.IOException;
import java.util.logging.Level;
import java.util.logging.Logger;
import io.grpc.stub.StreamObserver;
import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.ArrayList;
import java.net.InetAddress;

public class PlumConductor {
	private static final Logger logger = Logger.getLogger(PlumConductor.class.getName());
	private final Server server;
	private final int port;
	private final String host;
	private ArrayList<IPAddress> addressBook;

	public PlumConductor() {
		this("localhost", 50055);
	}

	public PlumConductor(String host, int port) {
		this.host = host;
		this.port = port;
		addressBook = new ArrayList<IPAddress>();
		server = ServerBuilder.forPort(port).addService(new PlumConductorServiceImpl(this)).build();
	}

	private void start() throws IOException {
		server.start();
		logger.info("Server started, listening on " + port);
		Runtime.getRuntime().addShutdownHook(new Thread(() -> {
			System.err.println("*** shutting down gRPC server since JVM is shutting down");
			PlumConductor.this.stop();
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

	// utils
	public void clearAddressBook() {
		this.addressBook.clear();
	}

	// entry point of PlumConductor
	public static void main(String[] args) {

		// run server on a thread. start and forgot
		Runnable serverTask = () -> {
			final PlumConductor server;
			server = new PlumConductor();
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

		// check addressbook regularly, if it is not working, delete it from arrayList
	}

	static class PlumConductorServiceImpl extends PlumServiceGrpc.PlumServiceImplBase {
		private static final Logger logger = Logger.getLogger(PlumConductorServiceImpl.class.getName());

		private PlumConductor thisConductor;

		public PlumConductorServiceImpl(PlumConductor thisConductor) {
			this.thisConductor = thisConductor;
		}

		// is alive?
		@Override
		public void sayHello(HelloRequest req, StreamObserver<HelloReply> responseObserver) {
			HelloReply reply = HelloReply.newBuilder().setMessage("Hello!").build();
			responseObserver.onNext(reply);
			responseObserver.onCompleted();
		}

		// networking
		@Override
		public void getIP(CommonRequest req, StreamObserver<IPAddress> responseObserver) {
			try {
				InetAddress local = InetAddress.getLocalHost();
				String myIp = local.getHostAddress().toString();
				IPAddress address = IPAddress.newBuilder().setAddress(myIp).setPort(thisConductor.getPort()).build();
				responseObserver.onNext(address);
			} catch (Exception e) {
				System.err.println(e.getMessage());
			}
			responseObserver.onCompleted();
		}

		// address which peer have related features
		@Override
		public void addAddress(IPAddress req, StreamObserver<CommonResponse> responseObserver) {
			// get address from client request
			IPAddress addressToSet = req;
			logger.info("set Address: " + addressToSet.getAddress() + ":" + addressToSet.getPort());

			// add address to peer's address book
			// get addressbook from connected peer
			ArrayList<IPAddress> peerAddressBook = thisConductor.getAddressBook();
			CommonResponse res;

			// check duplication
			if (!peerAddressBook.contains(addressToSet)) {
				thisConductor.getAddressBook().add(addressToSet);
				res = CommonResponse.newBuilder().setSuccess(true).build();
			} else {
				String errorMsg = "ERROR: DUPLICATED";
				res = CommonResponse.newBuilder().setSuccess(false).setError(errorMsg).build();
				logger.log(Level.INFO, "address duplicated. do nothing");
			}

			// response to client
			responseObserver.onNext(res);
			responseObserver.onCompleted();
		}

		@Override
		public StreamObserver<IPAddress> setAddressBook(final StreamObserver<CommonResponse> responseObserver) {
			return new StreamObserver<IPAddress>() {
				@Override
				public void onNext(IPAddress address) {
					// get address from client request
					IPAddress addressToSet = address;
					logger.info("set Address: " + addressToSet.getAddress());

					// get addressbook from connected peer
					ArrayList<IPAddress> peerAddressBook = thisConductor.getAddressBook();

					// check duplication
					if (!peerAddressBook.contains(addressToSet)) {
						thisConductor.getAddressBook().add(addressToSet);
					} else {
						// do nothing
						logger.log(Level.INFO, "address duplicated. do nothing");
					}
				}

				@Override
				public void onError(Throwable t) {
					logger.log(Level.WARNING, "setAddresssBook cancelled");
				}

				@Override
				public void onCompleted() {
					CommonResponse res = CommonResponse.newBuilder().setSuccess(true).build();
					responseObserver.onNext(res);
					responseObserver.onCompleted();
				}
			};
		}

		@Override
		public void getAddressBook(CommonRequest req, StreamObserver<IPAddress> responseObserver) {
			ArrayList<IPAddress> addressBook = thisConductor.getAddressBook();
			for (IPAddress address : addressBook) {
				responseObserver.onNext(address);
			}
			responseObserver.onCompleted();
		}

		@Override
		public void clearAddressBook(CommonRequest req, StreamObserver<CommonResponse> responseObserver) {
			CommonResponse res = CommonResponse.newBuilder().setSuccess(true).build();
			
			thisConductor.clearAddressBook();
			
			responseObserver.onNext(res);
			responseObserver.onCompleted();
		}
	}
}