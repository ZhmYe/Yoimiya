# todo 
import FigureEngine
import generate_data_json as generate_data_json
ROOT = "/root/Yoimiya/logWriter"
if __name__ == "__main__":
    engine = FigureEngine.FiguireEngine(ROOT)
    generate_data_json.process(ROOT + "/log/N-Split-nbLoop_Test")
    outputs = engine.read_data_from_json("log/N-Split-nbLoop_Test")
    for output in outputs:
        engine.draw(output, output["name"])