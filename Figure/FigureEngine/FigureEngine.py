# 这里统一写画图的函数，包括读取的文件路径
class FiguireEngine:
    # 这里考虑每个test的文件所要读取的东西不同，因此考虑在每个test里写一个python文件运行后得到输出，输出到某个文件output.json下
    def __init__(self, path):
        self.root = ""
        self.path = path # 要读取的所有文件所在的路径,在该路径下会有一个output.json
    def read_data_from_json(self):
        
