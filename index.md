---
title:  "Personal site for articles, useful links, and files"
---
<div class="row">
    <!-- Segunda Coluna: Categoria Soft Skills -->
  <div class="col-md-6">            
    {% assign categoria2 = "hard-skills" %}
    {% assign posts_hs = site.categories[categoria2] | sort: "date" | reverse %}
    <div class="card">
      <div class="card-body">
        <h5 class="card-title">Hard Skills</h5>              
        <p class="card-text">Últimos posts na categoria</p>
      </div>
      <ul class="list-group list-group-flush">
      {% for post in posts_hs limit:5 %}
        <li class="list-group-item"><a href="{{ post.url }}">{{ post.title }}</a> <span class="badge badge-primary badge-pill">{{ post.date | date: "%d/%m/%Y" }}</span></li>
      {% endfor %}
      </ul>    
      <div class="card-footer bg-transparent text-right">
          <a href="categories/hard-skills"> Mais artigos...</a>
      </div>      
    </div>
  </div>

  <!-- Segunda Coluna: Categoria Soft Skills -->
  <div class="col-md-6">            
    {% assign carreira = "carreira" %}
    {% assign posts_hs = site.categories[carreira] | sort: "date" | reverse %}
    <div class="card">
      <div class="card-body">
        <h5 class="card-title">Carreira</h5>              
        <p class="card-text">Últimos posts na categoria</p>
      </div>
      <ul class="list-group list-group-flush">
      {% for post in posts_hs limit:5 %}
        <li class="list-group-item"><a href="{{ post.url }}">{{ post.title }}</a> <span class="badge badge-primary badge-pill">{{ post.date | date: "%d/%m/%Y" }}</span></li>
      {% endfor %}
      </ul>    
      <div class="card-footer bg-transparent text-right">
          <a href="categories/carreira"> Mais artigos...</a>
      </div>      
    </div>
  </div>
</div>

