using System.Net;
using System.Text;
using Microsoft.AspNetCore.Components;
using Microsoft.FluentUI.AspNetCore.Components;
using Icons = Microsoft.FluentUI.AspNetCore.Components.Icons;

namespace GreenerBlazor.Services;

public class ExceptionService(IDialogService dialogService, IMessageService messageService)
{
    public async Task<bool> HandleAsMessageBox(HttpRequestException exc)
    {
        var (title, markupMessage) = CreateErrorMessage(exc);

        var dialog = await dialogService.ShowMessageBoxAsync(
            new DialogParameters<MessageBoxContent>
            {
                Content = new MessageBoxContent
                {
                    Title = title,
                    MarkupMessage = markupMessage,
                    Icon = new Icons.Regular.Size24.ErrorCircle(),
                    IconColor = Color.Error,
                },
                SecondaryAction = null,
            }
        );

        await dialog.Result;
        return true;
    }

    public async Task<bool> HandleAsMessageBar(HttpRequestException exc)
    {
        var (title, markupMessage) = CreateErrorMessage(exc);

        await messageService.ShowMessageBarAsync(options =>
        {
            options.Title = title;
            options.Intent = MessageIntent.Error;
            options.Section = "MESSAGES_TOP";
            options.AllowDismiss = false;
            options.Body = markupMessage.ToString();
        });
        return true;
    }

    public void ClearMessageBar() => messageService.Clear();

    private static (string title, MarkupString markupMessage) CreateErrorMessage(
        HttpRequestException exc
    )
    {
        var title = "Error";
        var message = exc.Message;

        if (exc.Data.Contains("StatusCode"))
        {
            var statusCode = (HttpStatusCode)exc.Data["StatusCode"]!;

            var isGenericMessage =
                string.IsNullOrEmpty(message)
                || message.Contains("Unauthorized")
                || message.Contains("Forbidden")
                || message.Contains("Bad Request")
                || message.Contains("Internal Server Error")
                || message.Contains("Not Found");

            if (isGenericMessage)
            {
                (title, message) = statusCode switch
                {
                    HttpStatusCode.Unauthorized => (
                        "Authentication Failed",
                        "Please log in again."
                    ),
                    HttpStatusCode.Forbidden => (
                        "Access Denied",
                        "You don't have permission to perform this action."
                    ),
                    HttpStatusCode.NotFound => (
                        "Not Found",
                        "The requested resource was not found."
                    ),
                    HttpStatusCode.BadRequest => (
                        "Invalid Request",
                        "Please check your input and try again."
                    ),
                    HttpStatusCode.InternalServerError => (
                        "Server Error",
                        "Server error occurred. Please try again later."
                    ),
                    HttpStatusCode.TooManyRequests => (
                        "Rate Limited",
                        "Too many requests. Please try again later."
                    ),
                    _ => ("Error", exc.Message),
                };
            }
            else
            {
                title = statusCode switch
                {
                    HttpStatusCode.Unauthorized => "Authentication Failed",
                    HttpStatusCode.Forbidden => "Access Denied",
                    HttpStatusCode.NotFound => "Not Found",
                    HttpStatusCode.BadRequest => "Invalid Request",
                    HttpStatusCode.InternalServerError => "Server Error",
                    HttpStatusCode.TooManyRequests => "Rate Limited",
                    _ => "Error",
                };
            }
        }

        var markupMessage = new MarkupString(
            $"<span style=\"white-space: pre-line\">{message}</span>"
        );

        return (title, markupMessage);
    }
}
