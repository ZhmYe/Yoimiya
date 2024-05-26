import os
import re
import json
import sys
def list2string(list):
    result = ""
    for item in list:
        result += (item + ", ")
    return result[:len(result) - 2]
def process(process_log):
    # 读取当前目录下的所有实验文件夹

    current_directory = "/root/Yoimiya/logWriter/log/MisalignedParalleling_nbTask_Test"

    # 列出当前目录中的所有文件夹
    directories = [d for d in os.listdir(current_directory) if os.path.isdir(os.path.join(current_directory, d))]

    print("{} has {} output dir: {}".format(current_directory, len(directories), list2string(directories)))
    # 遍历所有的测试文件夹
    output = {}
    for directory in directories:
        print("\t enter: {}".format(directory))
        if directory == "__pycache__":
            continue
        circuit_name = directory # 电路名
        output[circuit_name] = {}
        # 获取每个文件夹中的所有文件夹nbTask__{}
        directory_path = os.path.join(current_directory, directory)
        nb_task_directory = [d for d in os.listdir(directory_path) if os.path.isdir(os.path.join(directory_path, d))]
        for nb_task in nb_task_directory:
            if not nb_task.startswith("nbTask_"):
                print("\t\t {} not start with nbTask_, pass".format(nb_task))
                continue
            loop_number = int(nb_task[7:]) # 循环数
            output[circuit_name][loop_number] = {}
            loop_path = os.path.join(directory_path, nb_task)
            logs = [f for f in os.listdir(loop_path) if os.path.isfile(os.path.join(loop_path, f))]
            if len(logs) != 3:
                print("error: len(logs) != 3, in {}".format(nb_task))
                continue
            for log in logs:
                log_name = "misaligned_paralleling"
                if "misaligned_paralleling_cut_2" in log:
                    print("\t\t\t process misaligned_paralleling log: {}".format(os.path.join(loop_path, log)))
                    log_name = "2_split"
                elif "misaligned_paralleling_cut_3" in log:
                    print("\t\t\t process misaligned_paralleling log: {}".format(os.path.join(loop_path, log)))
                    log_name = "3_split"
                elif "serial_running" in log:
                    print("\t\t\t process serial_running log: {}".format(os.path.join(loop_path, log)))
                    log_name = "serial_running"
                else:
                    print("\t\t\t error: error log format!!!")
                    continue
                log_output = process_log(os.path.join(loop_path, log), log_name=log_name)
                output[circuit_name][loop_number][log_name] = log_output
    # print(output)
    with open(os.path.join(current_directory, "data.json"), "w", encoding="utf-8") as f:
        json.dump(output, f)
        f.close()
# generate_data_json()