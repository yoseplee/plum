package playground.rmi;
        
import java.rmi.registry.Registry;
import java.rmi.registry.LocateRegistry;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
        
public class Server1 implements Hello {
        
    public Server1() {}

    public String sayHello() {
        return "Hello, world! from #1";
    }
        
    public static void main(String args[]) {
        
        try {
            Server1 obj = new Server1();
            Hello stub = (Hello) UnicastRemoteObject.exportObject(obj, 0);

            // Bind the remote object's stub in the registry
            Registry registry = LocateRegistry.getRegistry(1109);
            registry.bind("Hello1", stub);

            System.err.println("Server1 ready");
        } catch (Exception e) {
            System.err.println("Server1 exception: " + e.toString());
            e.printStackTrace();
        }
    }
}