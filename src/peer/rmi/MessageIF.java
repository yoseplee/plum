package peer.rmi;

import java.rmi.Remote;
import java.rmi.RemoteException;

public interface MessageIF extends Remote {
    String sayHello() throws RemoteException;
    void sendMessage(String message) throws RemoteException;
    String getIP() throws RemoteException;
}
