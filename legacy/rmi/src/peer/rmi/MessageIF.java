package peer.rmi;

import java.rmi.Remote;
import java.rmi.RemoteException;
import java.util.ArrayList;

public interface MessageIF extends Remote {
    String sayHello() throws RemoteException;
    void sendMessage(String message) throws RemoteException;
    String getIP() throws RemoteException;
    String setAddressBook(ArrayList<String> addressBook) throws RemoteException;
    String addAddress(String address) throws RemoteException;
    String printAllAddressBook() throws RemoteException;
    String addTransaction(String message) throws RemoteException;
    String printMempool() throws RemoteException;
}
