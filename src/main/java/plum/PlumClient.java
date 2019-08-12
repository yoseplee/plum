package plum;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.stub.StreamObserver;
import plum.PlumServiceGrpc.PlumServiceBlockingStub;
import plum.PlumServiceGrpc.PlumServiceStub;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.logging.Level;
import java.util.logging.Logger;

public class PlumClient {
	private static final Logger logger = Logger.getLogger(PlumClient.class.getName());
	private final ManagedChannel channel;
	private final PlumServiceBlockingStub blockingStub;
	private final PlumServiceStub asyncStub;

	// initializer suite
	public PlumClient(String host, int port) {
		this(ManagedChannelBuilder.forAddress(host, port).usePlaintext());
	}

	public PlumClient(ManagedChannelBuilder<?> channelBuilder) {
		channel = channelBuilder.build();
		blockingStub = PlumServiceGrpc.newBlockingStub(channel);
		asyncStub = PlumServiceGrpc.newStub(channel);
	}

	// peer connection related features
	public void shutdown() throws InterruptedException {
		channel.shutdown().awaitTermination(5, TimeUnit.SECONDS);
	}

	// client-side rpc call implementation
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
		logger.info("IP: " + res.getAddress() + ":" + res.getPort());
	}

	public void addAddress(String address, int port) {
		logger.info("setting a address into connected peer");
		IPAddress req = IPAddress.newBuilder().setAddress(address).setPort(port).build();
		CommonResponse res;
		boolean isSuccess = false;
		try {
			res = blockingStub.addAddress(req);
			isSuccess = res.getSuccess();
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
			return;
		}
		
		logger.info("address set done(" + isSuccess + ")");
	}

	public void setAddressBook(ArrayList<IPAddress> addressBook) throws InterruptedException {
		logger.info("Setting addressbook into connected peer");
		final CountDownLatch latch = new CountDownLatch(1);
		StreamObserver<Empty> responseObserver = new StreamObserver<Empty>() {
			@Override
			public void onNext(Empty empty) {
				// do nothing
			}

			@Override
			public void onError(Throwable t) {
				logger.log(Level.WARNING, "RPC failed: {0}", Status.fromThrowable(t));
				latch.countDown();
			}

			@Override
			public void onCompleted() {
				logger.info("finish setAddressBook");
				latch.countDown();
			}
		};

		StreamObserver<IPAddress> requestObserver = asyncStub.setAddressBook(responseObserver);
		try {
			// loop
			for (IPAddress address : addressBook) {
				requestObserver.onNext(address);
			}
		} catch (RuntimeException e) {
			requestObserver.onError(e);
			throw e;
		}

		// notify completed
		requestObserver.onCompleted();

		if (!latch.await(1, TimeUnit.MINUTES)) {
			logger.warning("setAddressBook can not finish within 1 minutes");
		}
	}

	public ArrayList<IPAddress> getAddressBook() {
		ArrayList<IPAddress> addressBook = new ArrayList<IPAddress>();
		Empty req = Empty.newBuilder().build();

		Iterator<IPAddress> addresses;
		try {
			addresses = blockingStub.getAddressBook(req);
			for (int i = 1; addresses.hasNext(); i++) {
				IPAddress addressToSet = addresses.next();
				logger.info("[Client] Setting address: " + addressToSet);
				addressBook.add(addressToSet);
			}
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
		}
		return addressBook;
	}

	public void addTransaction(Transaction transaction) {
		logger.info("add transaction to connected peer");
		TransactionResponse res;
		try {
			res = blockingStub.addTransaction(transaction);
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
			return;
		}
		logger.info("transaction add done");
	}

	public ArrayList<Transaction> getMemPool() {
		ArrayList<Transaction> memPool = new ArrayList<Transaction>();
		Empty req = Empty.newBuilder().build();

		Iterator<Transaction> transactions;
		try {
			transactions = blockingStub.getMemPool(req);
			for(int i = 1; transactions.hasNext(); ++i) {
				Transaction tx = transactions.next();
				memPool.add(tx);
			}
		} catch (StatusRuntimeException e) {
			logger.log(Level.WARNING, "RPC failed: {0}", e.getStatus());
		}
		return memPool;
	}

	// entry point of client
	public static void main(String[] args) throws Exception {
		PlumClient client = new PlumClient("localhost", 50051);
		try {
			client.sayHello("HI");
			client.getIP();

			client.addAddress("192.168.33.2", 50051);
			client.addAddress("192.168.33.2", 50051);
			client.addAddress("192.168.33.2", 50051);
			client.addAddress("192.168.33.2", 50051);
			client.addAddress("192.168.33.2", 50051);

			ArrayList<IPAddress> tempAddressBook = new ArrayList<IPAddress>();
			tempAddressBook.add(IPAddress.newBuilder().setAddress("localhost").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.33").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.33").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.35").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.35").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.35").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.35").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.22").setPort(50051).build());
			tempAddressBook.add(IPAddress.newBuilder().setAddress("192.168.0.22").setPort(50051).build());
			client.setAddressBook(tempAddressBook);

			System.out.println("Getting addressbook from peer");
			ArrayList<IPAddress> book = client.getAddressBook();
			for(IPAddress temp : book) {
				System.out.println("address:: " + temp.getAddress() + ":" + temp.getPort());
			}

			System.out.println("Setting transaction into peer");
			client.addTransaction(Transaction.newBuilder().setTransaction("hello").build());
			client.addTransaction(Transaction.newBuilder().setTransaction("world").build());
			client.addTransaction(Transaction.newBuilder().setTransaction("this").build());
			client.addTransaction(Transaction.newBuilder().setTransaction("is").build());
			client.addTransaction(Transaction.newBuilder().setTransaction("yosep").build());

			System.out.println("Getting transaction from peer");
			ArrayList<Transaction> memPool = client.getMemPool();
			for(Transaction temp : memPool) {
				System.out.println("transaction from peer mempool: " + temp.getTransaction());
			}
		} finally {
			client.shutdown();
		}
	}
}
