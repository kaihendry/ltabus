function countdown(id, time) {
  if (!id) {
    return;
  }
  // console.log(id,time);
  var seconds = time / 1000;
  if (Math.abs(seconds) > 60) {
    id.innerHTML = parseInt(seconds / 60) + "m";
  } else {
    id.innerHTML = parseInt(seconds) + "s";
  }
  setTimeout(countdown, 1000, id, time - 1000);
}

window.addEventListener(
  "load",
  function () {
    var timings = document.getElementsByTagName("time");
    var now = new Date();
    for (let i = 0; i < timings.length; i++) {
      var arr = new Date(timings[i].getAttribute("datetime"));
      var elapsed = arr.getTime() - now.getTime();
      countdown(timings[i], elapsed);
    }
    var lastupdated = document.getElementById("lastupdated");
    countdown(lastupdated, Date.now() - now);

    var slog = JSON.parse(window.localStorage.getItem("history")) || {};

    var busstopcode = document.getElementById("id").value;
    var busstopname = document.getElementById("namedBusStop").innerHTML;

    console.log("DEBUG", busstopcode, busstopname);

    if (busstopcode) {
      if (typeof slog[busstopcode] === "undefined") {
        slog[busstopcode] = {};
        slog[busstopcode].count = 0;
        slog[busstopcode].name = busstopname;
      }
      try {
        slog[busstopcode].count++;
        slog[busstopcode].name = busstopname;
      } catch (e) {
        console.log(e);
      }

      window.localStorage.setItem("history", JSON.stringify(slog));
    } else {
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(function (position) {
          var lat = position.coords.latitude;
          var lng = position.coords.longitude;
          window.location = "/closest?lat=" + lat + "&lng=" + lng;
        });
      }
    }

    var sortable = [];
    for (var station in slog) {
      sortable.push([station, slog[station]]);
    }
    sortable.sort(function (a, b) {
      return a[1].count - b[1].count;
    });
    // console.debug(sortable);
    var ul = document.getElementById("stations");
    for (let i = sortable.length - 1; i >= 0; i--) {
      var key = sortable[i][0];
      var value = sortable[i][1];
      // console.log(key, value);

      var li = document.createElement("li");
      var link = document.createElement("a");
      if (value.name) {
        link.setAttribute(
          "href",
          "/?id=" + key + "&name=" + encodeURI(value.name)
        );
        link.appendChild(
          document.createTextNode(
            key + " " + value.name + " (" + value.count + ")"
          )
        );
      } else {
        link.setAttribute("href", "/?id=" + key);
        link.appendChild(document.createTextNode(key));
      }
      li.appendChild(link);
      ul.appendChild(li);
    }
  },
  false
);
