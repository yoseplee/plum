package peer.rmi;

import java.util.Scanner;
import java.util.ArrayList;

public class Cli {

    public Cli() {}

    public static void help() {
        System.out.println("CLI Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "help", "print all the commands"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "connect", "connect to specific peer. type ip address(v4)"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "list", "list up all the clients where connection have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "use", "select a client connected with a peer"));;
        
    }

    public static void clientHelp() {
        System.out.println("Client Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "sayHello", "send say hello to connect peer"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "getIP", "get IP address from connected peer, and add it to client's local addressbook"));
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addAddress", "send a single address to connected peer so that it can add it to its addressbook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "setAddressbook", "send a list of address to connected peer so that it can add it to its addressbook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "printAddressbook", "print all the addressess connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "exit", "quit client"));;
    }

    public static void printAllClients(ArrayList<Client> clientList) {
        if(clientList.size()==0) {
            System.err.println("Client List Empty");
            return;
        }
        int i = 0;
        System.out.println(String.format("%-5s%-16s%-10s", "idx", "|ip", "|port"));
        for(Client tmpClient : clientList) {
            System.out.println(String.format("%-5d%-16s%-10s", ++i, tmpClient.getHost(), tmpClient.getPort()));
        }
    }

    public static void welcomeCli() {
        System.out.println("===================");
        System.out.println("WELCOME TO PLUM CLI!");
        System.out.println("===================");
        System.out.println("help command for listing up commands that client have");
    }

    public static void welcomePeer() {
        System.out.println("*******************");
        System.out.println("WELCOME TO PLUM PEER!");
        System.out.println("*******************");
        System.out.println("help command for listing up commands that client have");
    }

    public static void main(String[] args) {
        ArrayList<Client> clientList = new ArrayList<Client>();
        Client currentClient = null;

        // String host = (args.length < 1) ? null : args[0];

        // Client client;
        Scanner scan = new Scanner(System.in);
        ArrayList<String> addressbook = new ArrayList<String>();
        welcomeCli();
        while(true) {
            System.out.print("&> ");
            String command = "";
            command = scan.nextLine();
            switch(command) {
                case "help":
                    help();
                    break;
                case "connect":
                    System.out.print("Which host(v4)? &> ");
                    String host = scan.nextLine();
                    System.out.println("try to connect to host: " + host);
                    Client client = new Client(host);
                    if(client.connect()) {
                        addressbook.add(client.getHost());
                        clientList.add(client);
                        System.out.println("connected to " + host);
                    } else {
                        System.err.println("connect error!");
                    }
                    break;           
                case "list":
                    printAllClients(clientList);
                    break;
                case "use":
                    int clientIdx = -1;
                    printAllClients(clientList);
                    System.out.print("which client to use? (idx) &> ");
                    clientIdx = Integer.parseInt(scan.nextLine());
                    //check bound
                    if(clientIdx > clientList.size() || clientIdx < 0) {
                        System.err.println("client idx out of bound!");
                        continue;
                    }
                    currentClient = clientList.get(clientIdx-1);
                    System.out.println(currentClient);
                    welcomePeer();
                    while(true) {
                        String response = "";
                        System.out.printf("%10s | &> ", currentClient.getHost());
                        String message = scan.nextLine();
                        //send RMI call
                        switch(message) {
                            case "help":
                                clientHelp();
                                break;
                            case "sayHello":
                                response = currentClient.sayHello();
                                break;
                            case "getIP":
                                response = currentClient.getIP();
                                break;
                            case "addAddress":
                                String address = "";
                                System.out.println("which address to add? &> ");
                                address = scan.nextLine();
                                response = currentClient.addAddress(address);
                                break;
                            case "setAddressbook":
                                //set client addressbook from cli
                                currentClient.setAddressbook(addressbook);                     
                                //set peer addressbook from client
                                response = currentClient.addAddressbook();
                                break;
                            case "printAddressbook":
                                response = currentClient.printAddressbook();
                                break;
                            default:
                                break;
                        }
                        if(message.equals("exit")) {
                            System.out.println("..back to CLI");
                            break;
                        }
                        //print
                        if(!response.isEmpty()) System.out.println("response: " + response);
                    }
                    welcomeCli();
                    break;
                case "exit":
                    System.out.println("Bye");
                    scan.close();
                    return;
                default:
                    break;
            }
        }
    }
}