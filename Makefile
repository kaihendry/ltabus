templates/index.html: templates/index.html.in
	npx critical --inline $< > $@

clean:
	rm -f templates/index.html
