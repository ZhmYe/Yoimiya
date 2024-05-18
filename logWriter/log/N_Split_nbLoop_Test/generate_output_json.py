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
# 处理的总逻辑
def process(dir):
    with open("{}/data.json".format(dir), encoding="utf-8") as f:
        output = json.load(f)
        f.close()
    # 这里先只处理内存部分
    figure_output = []
    # 内存柱状图
    for circuit in output:
        # 遍历每种电路的数据
        memory_figure = {
            "figure_type": "bar",
            "title": "Memory Reduce In N-Split {} Circuit With Different Loop Number".format(circuit),
            "x_label": "Loop Number",
            "y_label": "Memory(GB)",
            "name": "N-Split-nbLoop_Test-{}.png".format(circuit),
            "x_ticks": [],
            "data": {}
        }
        circuit_data = output[circuit]
        if len(circuit_data) == 0:
            continue
        sorted_key = sorted(circuit_data.keys(), key=int)
        memory_figure["x_ticks"] = sorted_key
        for key in sorted_key:
            for data in circuit_data[key]:
                if data not in memory_figure["data"]:
                    memory_figure["data"][data] = []
                memory_figure["data"][data].append(circuit_data[key][data]["memory"])
        figure_output.append(memory_figure)
    # todo 这里其实可以和上面的合在一起的
        
    # 内存提升折线图
    for circuit in output:
        memory_percent_figure = {
            "figure_type": "line",
            "title": "Memory Reduce Percent In N-Split {} Circuit with Different Loop Number".format(circuit),
            "x_label": "Loop Number",
            "y_label": "Percent(%)",
            "name": "N-Split-Percent-nbLoop_Test-{}".format(circuit),
            "x_ticks": [],
            "data": []
        }
        line_data = {"data": [], "legend": "memory Reduce Percent"}
        circuit_data = output[circuit]
        if len(circuit_data) == 0:
            continue
        sorted_key = sorted(circuit_data.keys(), key=int)
        memory_percent_figure["x_ticks"] = sorted_key
        for key in sorted_key:
            normal_running = circuit_data[key]["normal_running"]
            n_split = circuit_data[key]["n_split"]
            if normal_running["memory"] == 0:
                line_data["data"].append(0)
            else:
                line_data["data"].append((1 - n_split["memory"] / normal_running["memory"]) * 100)
        memory_percent_figure["data"].append(line_data)
        figure_output.append(memory_percent_figure)
    # print(figure_output)
    with open("{}/output.json".format(dir), "w", encoding="utf-8") as f:
        json.dump(figure_output, f)
        f.close()
# process()