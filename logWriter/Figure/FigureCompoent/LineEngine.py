import numpy as np
from matplotlib import pyplot as plt
# 折线图
class LineEngine:
    def __init__(self, width=0.8) -> None:
        self.width = width
        print("\t Line Engine Init Success!!!")
    def draw(self, json_data, output_file=""):
        # 提取数据
        title = json_data["title"]
        x_label = json_data["x_label"]
        y_label = json_data["y_label"]
        x_ticks = json_data["x_ticks"]
        data_series = json_data["data"]
        x = np.arange(len(x_ticks))  # 数字刻度的数组
        # 创建图形和轴
        fig, ax = plt.subplots(figsize=(10, 6))

        # 绘制每一条数据曲线
        for series in data_series:
            y_data = series["data"]
            legend = series["legend"]
            ax.plot(x, y_data, label=legend)

        # 设置标题和标签
        ax.set_title(title)
        ax.set_xlabel(x_label)
        ax.set_ylabel(y_label)

        # 设置x轴刻度
        ax.set_xticks(x)
        ax.set_xticklabels([str(tick) for tick in x_ticks])

        # 显示图例
        ax.legend()

        # # 显示网格
        # ax.grid(True)

        # 显示图形
        plt.tight_layout()
        if output_file != "":
            plt.savefig(output_file)
        # return ax