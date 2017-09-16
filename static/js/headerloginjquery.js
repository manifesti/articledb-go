$(document).ready(function() {
  $("#login-nav").submit( function(event) {
    event.preventDefault();
    $.post("/login/", $("#login-nav").serialize(), function(result) {
      if (result == "1") {
        $("#headerloginmessage").empty().html("<div class=\"alert alert-warning\">We couldn't recognize your credentials, please try again.</div>");
      } else if (result == "2") {
        $("#headerloginmessage").empty().html("<div class=\"alert alert-success\">Successful login!</div>");
        setTimeout( function() {window.location = "/view/"} , 1000);
      }}, "json");
  });
});
