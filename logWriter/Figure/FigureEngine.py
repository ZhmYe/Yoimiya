import os
import json
from FigureCompoent import BarEngine, LineEngine, TwinxEngine, StackBarEngine
# 这里统一写画图的函数，包括读取的文件路径
class FiguireEngine:
    # 这里考虑每个test的文件所要读取的东西不同，因此考虑在每个test里写一个python文件运行后得到输出，输出到某个文件output.json下
    def __init__(self, root, output_dir="output"):
        self.root = root
        self.output_dir = os.path.join(os.getcwd(), output_dir)
        # self.path = path # 要读取的所有文件所在的路径,在该路径下会有一个output.json
        print("Figuire Engine Init Success...")
    def read_data_from_json(self, path):
        complete_path = os.path.join(self.root, path)
        with open(os.path.join(complete_path, "output.json"), encoding="utf-8") as f:
            outputs = json.load(f)
            f.close()
        return outputs
    def draw(self, output, output_file):
        figure_type = output["figure_type"]
        # 柱状图
        if figure_type == "bar":
            barEngine = BarEngine.BarEngine()
            return barEngine.draw(output, os.path.join(self.output_dir, output_file))
        if figure_type == "line":
            lineEngine = LineEngine.LineEngine()
            return lineEngine.draw(output, os.path.join(self.output_dir, output_file))
        if figure_type == "stack_bar":
            stack_bar_engine = StackBarEngine.StackBarEngine()
            return stack_bar_engine.draw(output, os.path.join(self.output_dir, output_file))
    # def combine_twinx(self, ax1, ax2, title, output_file):
    #     twinX = TwinxEngine.TwinXEngine()
    #     twinX.draw(ax1, ax2, title, output_file)