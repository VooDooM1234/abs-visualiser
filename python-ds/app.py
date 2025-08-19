import matplotlib.pyplot as plt
import pandas as pd
import numpy as np
import plotly.express as px

import csv

with open('../.testdata/CPI.csv', newline='') as csvfile:
    reader = csv.DictReader(csvfile)
    for row in reader:
        print(row['TIME_PERIOD'], row['OBS_VALUE'])
print(row)


fig = px.bar(x=["a", "b", "c"], y=[1, 3, 2])
fig.show()

plt.plot([1,2,3], [4,5,6])
plt.show()

print("Hello world")
