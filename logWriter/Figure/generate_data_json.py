# 生成output.json
# 格式
# {
    # "figure_type": "这里写要画的图的类型"
    # "datas": {
        # [
        #     {
        #             "data": "这里是数据",
        #             "legend": "这里是图例",
        #             ....
        #     }
        # ]
    # }
    # "title": "图的title"
    # "x_ticks": ["横轴刻度"]
    # "x_label": "x轴名称"
    # "y_label": "y轴名称"
    # "name": "输出图片名称"
# }
import os
import re
import json
def list2string(list):
    result = ""
    for item in list:
        result += (item + ", ")
    return result[:len(result) - 2]
def time2key(key):
    if key == "Compile Time":
        return "compile"
    if key == "Set Up Time":
        return "setUp"
    if key == "Solve Time":
        return "solve"
    if key == "Split Time":
        return "split"
    if key == "Build Time":
        return "build"
def convert_to_ms(time_str):
    # 使用正则表达式匹配时间字符串中的分钟和秒
    match = re.match(r'(\d+)m([\d.]+)s', time_str)
    if match:
        minutes = int(match.group(1))
        seconds = float(match.group(2))
        
        # 将分钟和秒转换为毫秒
        total_ms = (minutes * 60 * 1000) + (seconds * 1000)
        return total_ms
    else:
        raise ValueError("time format error")
def process_log(path, log_name):
    data_dict = {
        "log": log_name,
        "memory": 0, # 统一为GB
        "time": {
            # 统一为ms
            "compile": 0,
            "setUp": 0,
            "solve": 0,
            "split": 0,
            "build": 0
        }
    }
    with open(path, encoding="utf-8") as f:
        content = f.read()
        f.close()
        # 去除[Record]: 行
    content = re.sub(r'\[Record\]:\t*\n', '', content)
    
    # 使用正则表达式匹配每一行的键值对
    matches = re.findall(r'\[(.*?)\]:\t*(.*)', content)
    for match in matches:
            key, value = match
            key = key.strip()
            value = value.strip()
            if key == "Memory Used":
                if value.endswith("GB"):
                    data_dict["memory"] = float(value.replace('GB', ''))
                else:
                    print("\t\t\t\t error: memory should end with GB!!!")
            if key in ["Compile Time", "Set Up Time", "Solve Time", "Split Time", "Build Time"]:
                dict_key = time2key(key)
                if value.endswith("ms"):
                    data_dict["time"][dict_key] = float(value.replace('ms', ''))
                elif value.endswith("s"):
                    # todo 这里还有小时没考虑
                    if "m" in value: # 超过1分钟
                        data_dict["time"][dict_key] = convert_to_ms(value)
                    else:
                        data_dict["time"][dict_key] = float(value.replace('s', '')) * 1000
                else:
                    print("\t\t\t\t error: less than ms!!!")
    return data_dict
# 处理的总逻辑
def process(process_directory):
    # 读取当前目录下的所有实验文件夹

    current_directory = process_directory

    # 列出当前目录中的所有文件夹
    directories = [d for d in os.listdir(current_directory) if os.path.isdir(os.path.join(current_directory, d))]

    print("{} has {} output dir: {}".format(current_directory, len(directories), list2string(directories)))
    # 遍历所有的测试文件夹
    output = {}
    for directory in directories:
        print("\t enter: {}".format(directory))
        circuit_name = directory # 电路名 
        output[circuit_name] = {}
        # 获取每个文件夹中的所有文件夹nbLoop_{}
        directory_path = os.path.join(current_directory, directory)
        nb_loop_directory = [d for d in os.listdir(directory_path) if os.path.isdir(os.path.join(directory_path, d))]
        for nb_loop in nb_loop_directory:
            if not nb_loop.startswith("nbLoop_"):
                print("\t\t {} not start with nbLoop, pass".format(nb_loop))
                continue
            loop_number = int(nb_loop[7:]) # 循环数
            output[circuit_name][loop_number] = {}
            loop_path = os.path.join(directory_path, nb_loop)
            logs = [f for f in os.listdir(loop_path) if os.path.isfile(os.path.join(loop_path, f))]
            if len(logs) != 2:
                print("error: len(logs) != 2, in {}".format(nb_loop))
                continue
            for log in logs:
                log_name = "n_split"
                if "n_split" in log:
                    print("\t\t\t process n_split log: {}".format(os.path.join(loop_path, log)))
                    log_name = "n_split"
                elif "normal_running" in log:
                    print("\t\t\t process normal_running log: {}".format(os.path.join(loop_path, log)))
                    log_name = "normal_running"
                else:
                    print("\t\t\t error: error log format!!!")
                    continue
                log_output = process_log(os.path.join(loop_path, log), log_name=log_name)
                output[circuit_name][loop_number][log_name] = log_output
    # print(output)
    with open(os.path.join(process_directory, "data.json"), "w", encoding="utf-8") as f:
        json.dump(output, f)
        f.close()
ROOT = "/root/Yoimiya/logWriter"
process(ROOT + "/log/N-Split-nbLoop_Test")