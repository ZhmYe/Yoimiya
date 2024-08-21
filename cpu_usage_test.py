import re
with open("./cpu_usage.log") as f:
    data = f.read()
    f.close()

usr_values = re.findall(r'\d{2}:\d{2}:\d{2}\s+all\s+([\d\.]+)', data)
usr_values = [float(data) for data in usr_values]
print(usr_values)