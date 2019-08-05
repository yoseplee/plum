package plum;

import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.util.logging.Logger;

public class PlumPeer {

	private static final Logger logger = Logger.getLogger(PlumPeer.class.getName());

	  private Server server;

	  private void start() throws IOException {
		  int port = 50051;
		  server = ServerBuilder.forPort(port)
			.addService(new TestImpl())
			.build()
			.start();
		  logger.info("Server started, listening on " + port);
		  Runtime.getRuntime().addShutdownHook(new Thread() {
			  @Override
			  public void run() {
				  System.err.println("*** shutting down gRPC server since JVM is shutting down");
				  PlumPeer.this.stop();
				  System.err.println("*** server shut down");
			  }
		  });
	  }

	  private void stop() {
		  if(server != null) server.shutdown();
	  }

	  private void blockUntilShutdown() throws InterruptedException {
		  if(server != null) {
			  server.awaitTermination();
		  }
	  }

	  public static void main(String[] args) throws IOException, InterruptedException {
		  final PlumPeer server = new PlumPeer();
		  server.start();
		  server.blockUntilShutdown();
	  }

	  static class TestImpl extends TestGrpc.TestImplBase {
		  @Override
		  public void sayHello(HelloRequest req, StreamObserver<HelloReply> responseObserver) {
			  HelloReply reply = HelloReply.newBuilder().setMessage("Hello " + req.getName()).build();
			  responseObserver.onNext(reply);
			  responseObserver.onCompleted();
		  }
	  }
}
