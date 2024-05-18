import os
import re
import json
import MisalignedParalleling_nbTask_Test
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
    if key == "Total Time":
        return "total"
def convert_to_ms_only_minute(time_str):
    # 使用正则表达式匹配时间字符串中的分钟和秒
    match = re.match(r'(\d+)m([\d.]+)s', time_str)
    if match:
#         hours = int(match.group(1))
        minutes = int(match.group(1))
        seconds = float(match.group(2))

        # 将分钟和秒转换为毫秒
        total_ms = (minutes * 60 * 1000) + (seconds * 1000)
        return total_ms
    else:
        print(time_str)
        raise ValueError("time format error")
def convert_to_ms_with_hours(time_str):
    # 使用正则表达式匹配时间字符串中的分钟和秒
    match = re.match(r'(\d+)h(\d+)m([\d.]+)s', time_str)
    if match:
        hours = int(match.group(1))
        minutes = int(match.group(2))
        seconds = float(match.group(3))

        # 将分钟和秒转换为毫秒
        total_ms = (hours * 60 *60 * 1000) + (minutes * 60 * 1000) + (seconds * 1000)
        return total_ms
    else:
        print(time_str)
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
            if key in ["Compile Time", "Set Up Time", "Solve Time", "Split Time", "Build Time", "Total Time"]:
                dict_key = time2key(key)
                if value.endswith("ms"):
                    data_dict["time"][dict_key] = float(value.replace('ms', ''))
                elif value.endswith("s"):
                    # todo 这里还有小时没考虑
                    if "h" in value:
                        data_dict["time"][dict_key] = convert_to_ms_with_hours(value)
                    elif "m" in value: # 超过1分钟
                        data_dict["time"][dict_key] = convert_to_ms_only_minute(value)
                    else:
                        data_dict["time"][dict_key] = float(value.replace('s', '')) * 1000
                else:
                    print("\t\t\t\t error: less than ms!!!")
    return data_dict
def process(path):
    if path == "MisalignedParalleling_nbTask_Test":
        from MisalignedParalleling_nbTask_Test import generate_data_json, generate_output_json
#         import MisalignedParalleling_nbTask_Test
#         generate_data_json.process(process_log)
#         generate_output_json.process(os.path.join(os.getcwd(), path))
    if path == "N_Split_nbLoop_Test":
        from N_Split_nbLoop_Test import generate_data_json, generate_output_json
    if path == "N_Split_Test":
        from N_Split_Test import generate_data_json, generate_output_json
    generate_data_json.process(process_log)
    generate_output_json.process(os.path.join(os.getcwd(), path))
process("N_Split_Test")