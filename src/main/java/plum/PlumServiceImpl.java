package plum;

import java.util.ArrayList;
import java.util.logging.Level;
import java.util.logging.Logger;
import io.grpc.stub.StreamObserver;
import java.net.InetAddress;

public class PlumServiceImpl extends PlumServiceGrpc.PlumServiceImplBase {
    private static final Logger logger = Logger.getLogger(PlumServiceImpl.class.getName());

    private PlumPeer thisPeer;

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

    @Override
    public void addAddress(IPAddress req, StreamObserver<Empty> responseObserver) {
        // get address from client request
        IPAddress addressToSet = req;
        logger.info("set Address: " + addressToSet.getAddress());

        // add address to peer's address book
        // get addressbook from connected peer
        ArrayList<IPAddress> peerAddressBook = thisPeer.getAddressBook();

        // check duplication
        if(!peerAddressBook.contains(addressToSet)) {
            thisPeer.getAddressBook().add(addressToSet);    
        } else {
            logger.log(Level.INFO, "address duplicated. do nothing");
        }

        // response to client
        Empty res = Empty.newBuilder().build();
        responseObserver.onNext(res);
        responseObserver.onCompleted();
    }

    @Override
    public StreamObserver<IPAddress> setAddressBook(final StreamObserver<Empty> responseObserver) {
        return new StreamObserver<IPAddress>() {
            @Override
            public void onNext(IPAddress address) {
                // get address from client request
                IPAddress addressToSet = address;
                logger.info("set Address: " + addressToSet.getAddress());

                // get addressbook from connected peer
                ArrayList<IPAddress> peerAddressBook = thisPeer.getAddressBook();

                // check duplication
                if(!peerAddressBook.contains(addressToSet)) {
                    thisPeer.getAddressBook().add(addressToSet);    
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
                Empty empty = Empty.newBuilder().build();
                responseObserver.onNext(empty);
                responseObserver.onCompleted();
            }
        };
    }

    @Override
    public void getAddressBook(Empty req, StreamObserver<IPAddress> responseObserver) {
        ArrayList<IPAddress> addressBook = thisPeer.getAddressBook();
        for(IPAddress address : addressBook) {
            responseObserver.onNext(address);
        }
        responseObserver.onCompleted();
    }
}