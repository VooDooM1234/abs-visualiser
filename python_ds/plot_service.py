

# from fastapi import FastAPI
# from fastapi.responses import HTMLResponse, JSONResponse
# import plotly.express as px
# import plotly.io as pio
# import pandas as pd

# app = FastAPI()
# db_conn = None  # This should be set externally if needed

# @app.get("/plot/test", response_class=HTMLResponse)
# async def get_plot():
#     fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
#     return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

# @app.get("/plot/test/json", response_class=JSONResponse)
# async def get_plot_json():
#     fig = px.line(x=[1, 2, 3], y=[10, 20, 15], title="Sample Line Plot")
#     return fig.to_dict()

# @app.get("/plot/bar-abs-cpi", response_class=HTMLResponse)
# async def get_abs_cpi_plot():
#     try:
#         query = "SELECT * FROM ABS_CPI"
#         df = pd.read_sql(query, db_conn)

#         df['value'] = pd.to_numeric(df['value'], errors='coerce').round(2).astype(float)

#         fig = px.bar(
#             df,
#             x="time_period",
#             y="value",
#             title="Consumer Price Index - Quarterly",
#             labels={'time_period': 'Time Period', 'value': 'CPI'}
#         )
#         return pio.to_html(fig, full_html=False, include_plotlyjs='cdn')

#     except Exception as e:
#         return HTMLResponse(content=f"<h3>Error generating plot: {e}</h3>", status_code=500)
