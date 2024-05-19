import numpy as np
from matplotlib import pyplot as plt
# 堆叠柱状图
class StackBarEngine:
    def __init__(self, width=0.2) -> None:
        self.width = width
        self.colors =  ['red', 'green', 'blue', 'orange', 'purple', 'grey', 'pink', 'cyan', 'magenta']
        print("\t Stack Bar Engine Init Success!!!")
    def draw(self, json_data, output_file=""):
        # 提取数据
        title = json_data["title"]
        x_label = json_data["x_label"]
        y_label = json_data["y_label"]
        x_ticks = json_data["x_ticks"]
        data = json_data["data"]
        bottom_values = [0] * len(x_ticks)
        x = np.arange(len(x_ticks))  # 数字刻度的数组
        # 创建图形和轴
        fig, ax = plt.subplots(figsize=(10, 6))
        index = 0
        data_list = []
        index = 0
        isBarh = False
        for key, values in data.items():
            if len(values) <= 4:
                isBarh = True
                plt.barh(x_ticks, values, label=key, color=self.colors[index], left=bottom_values, height=self.width, align='center')
            else:
                plt.bar(x_ticks, values, label=key, color=self.colors[index], bottom=bottom_values, width=self.width, align='center')
            bottom_values = [sum(x) for x in zip(bottom_values, values)]
            index+=1
        if isBarh:
            ax.set_title(title)
            ax.set_ylabel(x_label)
            ax.set_xlabel(y_label)
#             ax.set_yticks(x)
#             ax.set_yticklabels([str(tick) for tick in x_ticks])
        else:
            # 设置标题和标签
            ax.set_title(title)
            ax.set_xlabel(x_label)
            ax.set_ylabel(y_label)
            # 设置x轴刻度
#             ax.set_xticks(x)
#             ax.set_xticklabels([str(tick) for tick in x_ticks])

        # 显示图例
        ax.legend()

        # # 显示网格
        # ax.grid(True)

        # 显示图形
        plt.tight_layout()
        if output_file != "":
            plt.savefig(output_file)
        # return ax