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
    # 各部分时间堆叠柱状图
    for circuit in output:
        # 遍历每种电路的数据
        memory_figure = {
            "figure_type": "stack_bar",
            "title": "Time Structure In N-Split and Normal Running with {} Circuit".format(circuit),
            "x_label": "",
            "y_label": "Time(ms)",
            "name": "N-Split-Time_structure_Test-{}.png".format(circuit),
            "x_ticks": [],
            "data": {}
        }
        circuit_data = output[circuit]
        if len(circuit_data) == 0:
            continue
        # n_split / normal_running
        for key in circuit_data:
            memory_figure["x_ticks"].append(key)
            for data in circuit_data[key]["time"]:
                if data == "total":
                    continue
                if data not in memory_figure["data"]:
                    memory_figure["data"][data] = []
                memory_figure["data"][data].append(circuit_data[key]["time"][data])
        figure_output.append(memory_figure)
    # todo 这里其实可以和上面的合在一起的
    with open("{}/output.json".format(dir), "w", encoding="utf-8") as f:
        json.dump(figure_output, f)
        f.close()
# process()