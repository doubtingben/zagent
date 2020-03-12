#!/usr/bin/env python
# coding: utf-8

# In[2]:


import json
import urllib.request
url = "http://localhost:3001/public/zcashmetrics-653600-757716.json"
data = urllib.request.urlopen(url).read().decode()
obj = json.loads(data)
type(obj)
print(obj[0])
len(obj)


# In[5]:


import pandas as pd 
df = pd.DataFrame(obj)
df


# In[21]:


obj2 = obj[100:200]
len(obj2)
df2 = pd.DataFrame(obj2)


# In[24]:


from bokeh.plotting import figure
from bokeh.io import show, output_notebook
from bokeh.models import ColumnDataSource
source = ColumnDataSource(df)


# In[29]:


p = figure(title = "Value Pools", x_axis_label = "Block Height")
p.line(x='height', y='sapling_value_pool', source=source, color = 'red')
p.line(x='height', y='sprout_value_pool', source=source, color = 'blue')

output_notebook()
show(p)


# In[ ]:




