$(document).ready(function() {
  $("#loginForm").submit( function(event) {
    event.preventDefault();
    $.post("/login/", $("#loginForm").serialize(), function(result) {
      if (result == "1") {
        $("#loginmessage").empty().html("<div class=\"alert alert-warning\">We couldn't recognize your credentials, please try again.</div>");
      } else if (result == "2") {
        $("#loginmessage").empty().html("<div class=\"alert alert-success\">Successful login!</div>").delay(1000).window.location.replace("/view/");
      }}, "json");
    });
  });
