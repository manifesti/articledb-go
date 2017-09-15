$(document).ready(function() {
  $("#signupForm").submit( function(event) {
    event.preventDefault();
    $.post("/signup/", $("#signupForm").serialize(), function(result) {
      if (result == "1") {
        $("#signupmessage").empty().html("<div class=\"alert alert-warning\">Your password does not match the confirmation.</div>");
      } else if (result == "2") {
        $("#signupmessage").empty().html("<div class=\"alert alert-warning\">Your username is too long (maximum 32 characters).</div>");
      } else if (result == "3") {
        $("#signupmessage").empty().html("<div class=\"alert alert-warning\">Your email-address is malformed.</div>");
      } else if (result == "4") {
        $("#signupmessage").empty().html("<div class=\"alert alert-warning\">Error while inserting user to database.</div>");
      } else if (result == "5") {
        $("#signupmessage").empty().html("<div class=\"alert alert-success\">User has been added to database!</div>").delay(1000).window.location.replace("/view/");
      }}, "json");
    });
});
