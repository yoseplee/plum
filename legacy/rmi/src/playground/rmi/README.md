# rmi 서버 실행 가이드

자바의 클래스패스 설정에 의해 실행 위치에 영향을 받을 수 있음.

## 1. 컴파일
```bash
$ cd playground/rmi/
$ javac *.java
```

## 2. 서버 실행
```bash
$ cd ../../ #root directory of classpath
$ rmiresigrty &
# rmiregistry 1099 & (in case of running rmiregistry on specific port)
$ java playground.rmi.Server & # start server
```

## 3. 클라이언트 실행
```bash
$ java playground.rmi.Client # client send RMI message to the server
```