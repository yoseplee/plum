package plum;

import java.util.ArrayList;
import java.util.Random;
import java.util.concurrent.CountDownLatch;
import java.util.logging.Level;
import java.util.logging.Logger;
import io.grpc.stub.StreamObserver;
import java.net.InetAddress;

public class PlumServiceImpl extends PlumServiceGrpc.PlumServiceImplBase {
    private static final Logger logger = Logger.getLogger(PlumServiceImpl.class.getName());
    private final int GOSSIPBOUND = 3;
    
    private PlumPeer thisPeer;

    public PlumServiceImpl(PlumPeer thisPeer) {
        this.thisPeer = thisPeer;
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
            IPAddress address = IPAddress.newBuilder().setAddress(myIp).setPort(thisPeer.getPort()).build();
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
        ArrayList<IPAddress> peerAddressBook = thisPeer.getAddressBook();
        CommonResponse res;

        // check duplication
        if(!peerAddressBook.contains(addressToSet)) {
            thisPeer.getAddressBook().add(addressToSet); 
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
                CommonResponse res = CommonResponse.newBuilder().setSuccess(true).build();
                responseObserver.onNext(res);
                responseObserver.onCompleted();
            }
        };
    }

    @Override
    public void getAddressBook(CommonRequest req, StreamObserver<IPAddress> responseObserver) {
        ArrayList<IPAddress> addressBook = thisPeer.getAddressBook();
        for(IPAddress address : addressBook) {
            responseObserver.onNext(address);
        }
        responseObserver.onCompleted();
    }

    // simple transaction and gossip
    @Override
    public void addTransaction(Transaction req, StreamObserver<TransactionResponse> responseObserver) {
        ArrayList<Transaction> thisPeerMemPool = thisPeer.getMemPool();
        TransactionResponse res = TransactionResponse.newBuilder().setSuccess("true").build();

        // step1: verify transaction
        // if(!verify()) return; // do nothing

        // step2: duplication check
        if(!thisPeerMemPool.contains(req)) {
            // step3: add this transaction(req) into mempool
            thisPeerMemPool.add(req);
            // step4: gossip this transaction
            gossip(req);
            responseObserver.onNext(res);
            responseObserver.onCompleted();
        }
        // is duplicated. do nothing
        return;
    }

    @Override
    public void getMemPool(CommonRequest req, StreamObserver<Transaction> responseObserver) {
        ArrayList<Transaction> memPool = thisPeer.getMemPool();
        for(Transaction transaction : memPool) {
            responseObserver.onNext(transaction);
        }
        responseObserver.onCompleted();
    }

    public void gossip(Transaction transaction) {
        ArrayList<IPAddress> addressBook = thisPeer.getAddressBook();

        // randomly select from addressBook
        Random random = new Random();
        for(int i = 0; i < GOSSIPBOUND; ++i) {
            int rndIdx = random.nextInt(addressBook.size());
            IPAddress clientAddress = addressBook.get(rndIdx);
            String addressToSend = clientAddress.getAddress();
            int portToSend = clientAddress.getPort();
            logger.log(Level.INFO, "[GOSSIP] send to " + addressToSend + ":" + portToSend + "||" + transaction.getTransaction());
            
            //send to that client in thread(gossip: fire and forget)
            new Thread(() -> {
                try {
                    PlumClient client = new PlumClient(clientAddress.getAddress(), 50051);    
                    client.addTransaction(transaction);
                } catch (Exception e) {
                    System.err.println("Gossip Error: " + e.getMessage());
                }
            });
        }        
    }
}