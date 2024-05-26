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
        n_split_directory = [d for d in os.listdir(directory_path) if os.path.isdir(os.path.join(directory_path, d))]
        for n_split in n_split_directory:
            if not n_split.endswith("_split"):
                if n_split != "normal_running":
                    print("\t\t {} not end with _split, pass".format(n_split))
                    continue
            if n_split == "normal_running":
                split_number = 1
            else:
                split_number = int(n_split[:len(n_split) - 6]) # 循环数
            output[circuit_name][split_number] = {}
            loop_path = os.path.join(directory_path, n_split)
            logs = [f for f in os.listdir(loop_path) if os.path.isfile(os.path.join(loop_path, f))]
            if len(logs) != 1:
                print("error: len(logs) != 1, in {}".format(n_split))
                continue
            log_output = process_log(os.path.join(loop_path, logs[0]), log_name=n_split)
            output[circuit_name][split_number] = log_output
    # print(output)
    with open(os.path.join(current_directory, "data.json"), "w", encoding="utf-8") as f:
        json.dump(output, f)
        f.close()
# process()