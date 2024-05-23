+++
title = '{{ replace .File.ContentBaseName "-" " " | title }}'
date = {{ .Date }}
featured_image = '{{ printf "/images/%s.webp" .File.ContentBaseName }}'
+++
