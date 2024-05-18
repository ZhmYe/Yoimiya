# todo 
import FigureEngine
ROOT = "/root/Yoimiya/logWriter"
if __name__ == "__main__":
    engine = FigureEngine.FiguireEngine(ROOT)
    outputs = engine.read_data_from_json("log/N_Split_nbLoop_Test")
    for output in outputs:
        engine.draw(output, output["name"])