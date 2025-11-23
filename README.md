# Tokly
## МПИТ Окружной этап
## Кейс: «Мониторинг состояния виброгасителей, изоляторов и траверсов»

### Requirement
 - OpenCV
 - Mysql
 - YOLO onnx model


### Build
**Download dependency**
```shell
go mod dounload
```

**Install OpnCV**

<p>Windows:</p>

```shell
cd $GOPATH/src/gocv.io/x/gocv@v0.42.0
win_build_opencv.cmd
```

<p>Linux:</p>

```shell
cd $GOPATH/src/gocv.io/x/gocv@v0.42.0
make install
```

**Run**
```shell
go run cmd/detector/main.go
```

### Example config/config.yaml
```yaml
debug: true

http:
  host: "127.0.0.1:8080"
  read_timeout_sec: 10
  write_timeout_sec: 10
  ssl_key_path: ""
  ssl_cert_path: ""

  handle_timeout_sec: 20



mysql:
  host: "127.0.0.1"
  port: 3306
  user: "root"
  schema: "Tokly"
  password: "pass"
  connect_timeout_sec: 10


yolo_model:
  model: "/model/model.onnx"
  model_config: "/model/config.yaml"
  model-seg: "/model/model-seg.onnx"
  model-model_seg_config: "/model/config_seg.onnx"


default_lap_config:
  vibration_damper: 0
  festoon_insulators: 0
  traverse: 0
  nest: 0
  safety_sign+: 0
  bad_insulator: 5
  damaged_insulator: 3
  polymer_insulators: 0

  sum: 20




images_path: "./images"
```

### Example model config.yaml
```yaml
class-list:
  - vibration_damper
  - festoon_insulators
  - traverse
  - nest
  - safety_sign+
  - bad_insulator
  - damaged_insulator
  - polymer_insulators

size:
  width: 640
  height: 640

conf-threshold: 0.5
NMS-threshold: 0.5
```

### Example segmentation model config.yaml
```yaml
size:
  width: 640
  height: 640

conf-threshold: 0.5
NMS-threshold: 0.7
```