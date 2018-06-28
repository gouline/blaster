var blaster = {
    messageState: {
        queue: [],
        count: 0
    },

    recipientsField: {
        onCreateToken: function(tf, e) {
            if (e.attrs.type === "usergroup") {
                e.preventDefault();
                
                var field = $(tf);
                blaster.getTokenfieldInput(field).val("");
    
                $.each(e.attrs.children, function(index, child) {
                    field.tokenfield("createToken", child);
                });
    
                return;
            }
            
            if (e.attrs.value === e.attrs.label) {
                var tokens = e.attrs.value.split("|");
                if (tokens.length == 2) {
                    e.attrs.value = tokens[0];
                    e.attrs.label = tokens[1];
                    e.attrs.type = "user";
                } else {
                    e.preventDefault();
                    return;
                }
            }
    
            e.attrs.value += "|" + e.attrs.label
    
            $.each($(tf).tokenfield("getTokens"), function(i, token) {
                if (blaster.getPipedValue(token.value) === blaster.getPipedValue(e.attrs.value)) {
                    e.preventDefault();
                    return;
                }
            });
        },

        onCreatedToken: function(tf, e) {
            if (!e.attrs.type) {
                $(e.relatedTarget).addClass("invalid");
            }
        },
    },

    getPipedValue: function(value) {
        return value.split("|")[0];
    },

    getTokenfieldInput: function(field) {
        return $("#" + $(field).attr("id") + "-tokenfield");
    },

    sendMessage: function() {
        var users = $("#recipients-field").val().split(", ");
        var message = $("#message-field").val();
        
        users = $.map(users, function(u, i) {
            return blaster.getPipedValue(u);
        });

        var missingUsers = users.length == 1 && users[0] == "";
        var missingMessage = message.length == 0;

        if (missingUsers || missingMessage) {
            $("#recipients-field").closest(".form-group").toggleClass("has-error", missingUsers);
            $("#message-field").closest(".form-group").toggleClass("has-error", missingMessage);
            return;
        }

        blaster.setFormEnabled(false);

        blaster.messageState.queue = [];
        blaster.messageState.count = users.length;
        for (user of users) {
            blaster.messageState.queue.push({user: user, message: message});
        }

        blaster.setProgressEnabled(true)
        blaster.nextMessage();
    },

    nextMessage: function() {
        var remaining = blaster.messageState.queue.length;

        blaster.setProgressValue(blaster.messageState.count - remaining, blaster.messageState.count);

        if (remaining == 0) {
            blaster.resetForm(true);
            return;
        }

        var message = blaster.messageState.queue.pop();

        $.ajax({
            type: "POST",
            url: "/api/send",
            data: JSON.stringify(message),
            contentType: "application/json; charset=utf-8",
            dataType: "json",
            success: function(data) {
                blaster.nextMessage();
            },
            error: function(data) {
                alert("Error sending message:\n" + JSON.stringify(data, null, 2));
                blaster.resetForm(false);
            }
        });
    },

    resetForm: function(success) {
        blaster.messageState.queue = [];
        blaster.messageState.count = 0;

        if (success) {
            $("#recipients-field").tokenfield('setTokens', []);
            $("#message-field").val("");
        } else {
            blaster.setProgressEnabled(false);
        }

        blaster.setFormEnabled(true);
    },

    resetFormErrors: function() {
        $("#recipients-field").closest(".form-group").toggleClass("has-error", false);
        $("#message-field").closest(".form-group").toggleClass("has-error", false);
    },

    setFormEnabled: function(enabled) {
        $("#recipients-field").tokenfield(enabled ? 'enable' : 'disable');
        $("#message-field").prop("disabled", !enabled);
        $("#submit-button").prop("disabled", !enabled);
    },

    setProgressEnabled: function(enabled) {
        $("#progress").show();
    },

    setProgressValue: function(current, total) {
        var percentage = current / total * 100;
        $("#progress-bar")
            .text(percentage < 100 ? current + "/" + total : "Complete!")
            .css("width", percentage + "%")
            .prop("aria-valuenow", percentage);
    }
};
