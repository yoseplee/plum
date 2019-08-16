package plum.cli;

import plum.*;
import java.util.Scanner;
import java.util.concurrent.CountDownLatch;
import java.util.ArrayList;

public class PlumCli {

    public static void help() {
        System.out.println("CLI Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "help", "print all the commands"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "connect", "connect to specific peer. type ip address(v4)"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "list", "list up all the clients where connection have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "use", "select a client connected with a peer"));;
        
    }

    public static void clientHelp() {
        System.out.println("PlumClient Function List");
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "sayHello", "send say hello to connect peer"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "getIP", "get IP address from connected peer, and add it to client's local addressBook"));
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addAddress", "send a single address to connected peer so that it can add it to its addressBook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "setAddressBook", "send a list of address to connected peer so that it can add it to its addressBook"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "printAddressBook", "print all the addressess connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "addTransaction", "add a transaction to peer and gossip"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "printMempool", "print all the mempool of connected peer have"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "clearAddressBook", "clear out address book of connected peer"));;
        System.out.println(String.format("%-4s%-20s|%-50s", "*--|", "exit", "quit client"));;

    }

    public static void printAllClients(ArrayList<PlumClient> clientList) {
        if(clientList.size()==0) {
            System.err.println("PlumClient List Empty");
            return;
        }
        int i = 0;
        System.out.println(String.format("%-5s%-16s%-10s", "idx", "|ip", "|port"));
        for(PlumClient tmpClient : clientList) {
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

    public static void closeAllConnections(ArrayList<PlumClient> clientList) {
        int size = clientList.size();
        System.out.println("close all connections(" + size + ")");
        if(size == 0) {
            System.out.println("PlumClient List Empty");
            return;
        }
        CountDownLatch latch = new CountDownLatch(size);
        for(PlumClient client : clientList) {
            new Thread(() -> {
                try {
                    client.shutdown();
                } catch (InterruptedException e) {
                    e.printStackTrace();
                }
                latch.countDown();
            }).start();
        }
        try {
            latch.await();
            System.out.println("all connections are closed");
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }

    public PlumCli() {}
    public static void main(String[] args) {
        ArrayList<PlumClient> clientList = new ArrayList<PlumClient>();
        PlumClient currentClient = null;

        // String host = (args.length < 1) ? null : args[0];

        // PlumClient client;
        Scanner scan = new Scanner(System.in);
        ArrayList<IPAddress> addressBook = new ArrayList<IPAddress>();
        welcomeCli();
        help();
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
                    System.out.print("Which port(default| peer:50051, conductor: 50055)? &> ");
                    int port = Integer.parseInt(scan.nextLine());
                    System.out.println("try to connect to host: " + host + ":" + port);
                    PlumClient client = new PlumClient(host, port);
                    addressBook.add(client.getIP());
                    clientList.add(client);
                    System.out.println("connected to " + host + ":" + port);
                    break;           
                case "list":
                    printAllClients(clientList);
                    break;
                case "use":
                    int clientIdx = -1;
                    printAllClients(clientList);
                    if(clientList.size() == 0) continue;
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
                    clientHelp();
                    while(true) {
                        String response = "";
                        System.out.printf("%10s:%5s | &> ", currentClient.getHost(), currentClient.getPort());
                        String message = scan.nextLine();
                        //send RMI call
                        switch(message) {
                            case "help":
                                clientHelp();
                                break;
                            case "sayHello":
                                // response = currentClient.sayHello();
                                currentClient.sayHello("cli");
                                break;
                            case "getIP":
                                response = currentClient.getIP().getAddress();
                                break;
                            case "addAddress":
                                System.out.println("which address to add(v4)? &> ");
                                String addressToAdd = scan.nextLine();
                                System.out.println("which port? &> ");
                                int portToAdd = 0;
                                try {
                                    portToAdd = Integer.parseInt(scan.nextLine());
                                } catch (Exception e) {
                                    System.err.println("UNEXPECTED: port number is integer.");
                                    e.printStackTrace();
                                    continue;
                                }
                                currentClient.addAddress(addressToAdd, portToAdd);
                                break;
                            case "setAddressBook":
                                //set client addressBook from cli
                                try {
                                    currentClient.setAddressBook(addressBook);                     
                                } catch (InterruptedException e) {
                                    e.printStackTrace();
                                }
                                //set peer addressBook from client
                                // response = currentClient.addAddressbook();
                                break;
                            case "printAddressBook":
                                ArrayList<IPAddress> book = currentClient.getAddressBook();
                                for(IPAddress temp : book) {
                                    System.out.println("address:: " + temp.getAddress() + ":" + temp.getPort());
                                }
                                break;
                            case "addTransaction":
                                String transaction = "";
                                System.out.println("which transaction to add? &> ");
                                transaction = scan.nextLine();
                                long currentTime = System.currentTimeMillis();
                                Transaction transactionToAdd = Transaction.newBuilder().setTransaction(transaction).setTime(currentTime).build();
                                currentClient.addTransaction(transactionToAdd);
                                break;
                            case "printMempool":
                                ArrayList<Transaction> memPool = currentClient.getMemPool();
                                for(Transaction temp : memPool) {
                                    // System.out.println("transaction from peer mempool: " + temp.getTransaction());
                                    System.out.printf("%10s | %10d\n", temp.getTransaction(), temp.getTime());
                                }
                                break;
                            case "clearAddressBook":
                                currentClient.clearAddressBook();
                                break;
                            default:
                                break;
                        }
                        if(message.equals("exit")) {
                            System.out.println("..back to CLI");
                            break;
                        }
                    }
                    welcomeCli();
                    help();
                    break;
                case "exit":
                    System.out.println("Bye");
                    scan.close();
                    // close all the peer connections?
                    closeAllConnections(clientList);
                    return;
                default:
                    break;
            }
        }
    }
}