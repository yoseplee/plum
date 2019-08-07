package plum;

import io.grpc.stub.StreamObserver;
import java.net.InetAddress;

public class PlumServiceImpl extends PlumServiceGrpc.PlumServiceImplBase {
    PlumPeer thisPeer;
    
    public PlumServiceImpl(PlumPeer thisPeer) {
        this.thisPeer = thisPeer;
    }

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