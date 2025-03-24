# 数据存储说明

## 根目录

所有数据存储在主机的 `/data/beagle-wind` 目录下：

```bash
/data/beagle-wind/
```

如/data/beagle-wind/lutris:/data/beagle-wind/
在主机上目录是/data/beagle-wind/lutris
在容器内部对应的目录是/data/beagle-wind/

## 游戏目录

游戏相关文件存储在 `/data/beagle-wind/games` 目录下：

```bash
/data/beagle-wind/games/
```

该目录用于存储所有游戏的安装文件、存档等数据。

### Mods

游戏的 Mods 文件通常直接解压到游戏目录中：

```bash
/data/beagle-wind/games/<game_name>/mods/
```

一般来讲客户想要下载 Mods，会直接解压至游戏目录。这样可以确保 Mods 能被游戏正确加载。
任何有 Mods 的游戏在加载前后都要考虑，先把 Mods 干掉。

### Saves

游戏存档的位置因游戏类型而异：

- Wine/Proton 游戏：存档通常位于 Proton 目录
- 原生游戏：存档可能直接位于游戏目录下

```bash
/data/beagle-wind/games/<game_name>/saves/
```

对于大部分 Wine 游戏设计来讲，存档一般在 Proton 目录里面，但是仍然有一部分游戏存档在游戏目录下。

## 数据卷

容器数据卷挂载在 `/data/beagle-wind/volumes` 目录下：

```bash
/data/beagle-wind/volumes/
```

该目录用于持久化存储容器运行时的数据。
