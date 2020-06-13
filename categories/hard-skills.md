---
title: Últimos artigos do blog da categoria hard-skills
---

{% assign categoria = "hard-skills" %}
{% assign posts = site.categories[categoria] | sort: "date" | reverse %}

<ul>
  {% for post in posts %}
    <li>
      <a href="{{ post.url }}">{{ post.title }}</a> <span class="badge badge-primary badge-pill">{{ post.date | date: "%d/%m/%Y" }}</span>
    </li>
  {% endfor %}
</ul>

