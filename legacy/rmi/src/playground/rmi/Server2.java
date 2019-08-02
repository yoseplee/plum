package playground.rmi;
        
import java.rmi.registry.Registry;
import java.rmi.registry.LocateRegistry;
import java.rmi.RemoteException;
import java.rmi.server.UnicastRemoteObject;
        
public class Server2 implements Hello {
        
    public Server2() {}

    public String sayHello() {
        return "Hello, world! from #2";
    }
        
    public static void main(String args[]) {
        
        try {
            Server2 obj = new Server2();
            Hello stub = (Hello) UnicastRemoteObject.exportObject(obj, 0);

            // Bind the remote object's stub in the registry
            Registry registry = LocateRegistry.getRegistry();
            registry.bind("Hello2", stub);

            System.err.println("Server2 ready");
        } catch (Exception e) {
            System.err.println("Server2 exception: " + e.toString());
            e.printStackTrace();
        }
    }
}