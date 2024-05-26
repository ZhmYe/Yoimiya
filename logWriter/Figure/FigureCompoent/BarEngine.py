import numpy as np
from matplotlib import pyplot as plt
# 柱状图
class BarEngine:
    def __init__(self, width=0.8) -> None:
        self.width = width
        print("\t BarEngine Init Success!!!")
    def draw(self, json_data, output_file=""):
        title = json_data["title"]
        x_label = json_data["x_label"]
        y_label = json_data["y_label"]
        x_ticks = json_data["x_ticks"]
        data = json_data["data"]
        width = self.width / len(data)  # 动态计算柱状图的宽度，确保所有柱子能并排显示
        # 创建X轴的位置
        x = np.arange(len(x_ticks))  # 数字刻度的数组
        # 创建子图
        fig, ax = plt.subplots(figsize=(10, 6))

        # 绘制柱状图
        rects = []
        for i, (key, values) in enumerate(data.items()):
            rects.append(ax.bar(x - (len(data.items()) -1) * width/2 + i*width, values, width, label=key))
        # 添加标题和标签
        ax.set_title(title)
        ax.set_xlabel(x_label)
        ax.set_ylabel(y_label)
        ax.set_xticks(x)
        ax.set_xticklabels(x_ticks)
        ax.legend()
        # 添加数值标签
        def autolabel(rects):
            for rect in rects:
                for bar in rect:
                    height = bar.get_height()
                    ax.annotate(f'{height:.2f}',
                                xy=(bar.get_x() + bar.get_width() / 2, height),
                                xytext=(0, 3),  # 3 points vertical offset
                                textcoords="offset points",
                                ha='center', va='bottom')

        autolabel(rects)
        # 显示图表
        if output_file != "":
            plt.savefig(output_file)
        # return ax