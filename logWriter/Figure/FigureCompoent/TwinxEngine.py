# import numpy as np
# from matplotlib import pyplot as plt
# # 公用y轴,这里目前仅支持柱状图和折线图
# class TwinXEngine:
#     def __init__(self, width=0.8):
#         self.width = width
#         print("\t TwinX Engine Init Success!!!")
#     def draw(self, json_data_1, json_data_2, title, output_file):
#         # 创建一个新的图形
#         # or not np.array_equal(ax1.get_xticklabels(), ax2.get_xticklabels()) 这里np.array_equal有问题，先不管
#         if ax1.get_xlabel() != ax2.get_xlabel() or not np.array_equal(ax1.get_xticks(), ax2.get_xticks()):
#             print("\t ax1's x != ax2's x")
#             return None 
#         print(ax1.get_xticklabels())
#         ax2_twin = ax1.twinx()  # 创建第二个Y轴
#         for line in ax2.get_lines():
#             ax2_twin.plot(line.get_xdata(), line.get_ydata(), color=line.get_color())

#         # 设置标签和标题
#         ax1.set_xlabel(ax1.get_xlabel())
#         ax1.set_ylabel(ax1.get_ylabel())
#         ax1.set_xticks(ax1.get_xticks())
#         ax1.set_xticklabels(ax1.get_xticklabels())
#         ax2_twin.set_ylabel(ax2.get_ylabel())
#         # ax.tick_params(axis='y', labelcolor='blue')
#         # ax2_twin.tick_params(axis='y', labelcolor='orange')

#         # 合并图例
#         lines, labels = ax1.get_legend_handles_labels()
#         lines2, labels2 = ax2_twin.get_legend_handles_labels()
#         ax1.legend(lines + lines2, labels + labels2, loc='upper left')

#         # 显示图形
#         plt.title(title)
#         if output_file != "":
#             plt.savefig(output_file)
#         return ax1