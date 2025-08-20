from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from fastapi.responses import JSONResponse
import plotly.express as px
import plotly.io as pio

app = FastAPI()

@app.get("/health")
def health_check():
    return {"status": "ok"}

# @app.get("/plot/test", response_class=HTMLResponse)
# def get_plot():

#     fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
    
#     html_graph = pio.to_html(fig, full_html=False, include_plotlyjs=False)
    
#     return html_graph

@app.get("/plot/test", response_class=HTMLResponse)
def get_plot():
    # Just return the container div
    return '<div id="plot-test"></div>'


@app.get("/plot/test/json", response_class=JSONResponse)
def get_plot_json():
    fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
    return fig.to_dict()
