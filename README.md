# Tokly

Мониторинг состояния виброгасителей, изоляторов и траверсов

### Requirement
 - OpenCV 4
 - Mysql
 - YOLO onnx model

### Add OpenCV pkgconfig to pach
```sh
set PATH = %PATH%;C:/opencv/build/install/x64/mingw/lib/pkgconfig/opencv4.pc
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
  model: "/model/best.onnx"
  model_config: "/model/config.yaml"


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

input: images
output: output0
size:
  width: 864
  height: 864

conf-threshold: 0.5
NMS-threshold: 0.5
```