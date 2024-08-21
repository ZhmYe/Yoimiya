#!/bin/bash

# 采样间隔（秒）
INTERVAL=0.01

# 总监控时间（秒），例如 20 分钟 = 20 * 60 = 1200 秒
TOTAL_TIME=130

# CPU 利用率输出文件
OUTPUT_FILE="cpu_usage.log"

# 清空输出文件
> $OUTPUT_FILE

#echo "时间, 用户空间 (%usr), 系统空间 (%sys), 空闲 (%idle)" > $OUTPUT_FILE

# 使用 timeout 控制监控时长
#timeout $TOTAL_TIME bash -c "
#while true; do
  mpstat 1 130 >> $OUTPUT_FILE

#  sleep $INTERVAL
#done
#"


echo "CPU利用率测试完成。结果已保存到 $OUTPUT_FILE"
