## ProcessとThreadの違い
- __Process__
  - 実行中のプログラム

- __Thread__
  - Processの中の実行単位(Flow)
  - 1つのProcessの中に1つ以上のThreadが存在する
    - 1つのProcessの中に2つ以上のThreadが存在する場合、`Multi Threading`という
    - 各Threadは同時に独立して動く
- Process間はMemory spaceを共有しない
  - 各Processは独立したMemory space上で動いて、他のProcessのMemory spaceには干渉できない
- 同じProcess内のThread間はMemory space(Heap)を共有する
  - ThreadsはProcessが持つMemory spaceしか使うことができない
- 参考URL
  - https://stackoverflow.com/questions/34689709/java-threads-and-number-of-cores
  - https://www.geeksforgeeks.org/difference-between-java-threads-and-os-threads/
  - https://www.youtube.com/watch?v=x-Lp-h_pf9Q&t=918s