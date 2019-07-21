# Dockerfile build

도커파일 빌드에 필요한 내용을 기록한다

# 중요!: 빌드 명령어의 시작 위치는 Server.java, Hello.java가 있는 ../../playground/rmi/에서 해야 한다
> ADD 명령어 때문. 추후 github 다운로드하는 형식으로 바꾼다면 수정의 여지가 있음

```bash
$ cd ../../playground/rmi/
$ docker build --tag peer:{current_version} -f ../../dockerfile/peer/Dockerfile .
```