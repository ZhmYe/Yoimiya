import os
import re
import json
def list2string(list):
    result = ""
    for item in list:
        result += (item + ", ")
    return result[:len(result) - 2]
# 处理的总逻辑
def process(process_log):
    # 读取当前目录下的所有实验文件夹

    current_directory = "/root/Yoimiya/logWriter/log/N_Split_Test"

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
        # 获取每个文件夹中的所有文件夹nbLoop_{}
        directory_path = os.path.join(current_directory, directory)
        logs = [f for f in os.listdir(directory_path) if os.path.isfile(os.path.join(directory_path, f))]
        if len(logs) != 2:
            print("error: len(logs) != 2, in {}".format(directory_path))
            continue
        for log in logs:
            log_name = "n_split"
            if "n_split" in log:
                print("\t\t\t process n_split log: {}".format(os.path.join(directory_path, log)))
                log_name = "n_split"
            elif "normal_running" in log:
                print("\t\t\t process normal_running log: {}".format(os.path.join(directory_path, log)))
                log_name = "normal_running"
            else:
                print("\t\t\t error: error log format!!!")
                continue
            log_output = process_log(os.path.join(directory_path, log), log_name=log_name)
            output[circuit_name][log_name] = log_output
    # print(output)
    with open(os.path.join(current_directory, "data.json"), "w", encoding="utf-8") as f:
        json.dump(output, f)
        f.close()
# process()