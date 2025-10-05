using Blazored.LocalStorage;
using GreenerBlazor;
using GreenerBlazor.Helpers;
using GreenerBlazor.Services;
using Microsoft.AspNetCore.Components.Authorization;
using Microsoft.AspNetCore.Components.Web;
using Microsoft.AspNetCore.Components.WebAssembly.Hosting;
using Microsoft.FluentUI.AspNetCore.Components;
using Microsoft.FluentUI.AspNetCore.Components.DataGrid.Infrastructure;

var builder = WebAssemblyHostBuilder.CreateDefault(args);
builder.RootComponents.Add<App>("#app");
builder.RootComponents.Add<HeadOutlet>("head::after");

var apiBaseUrl =
    builder.Configuration["GREENER:HOST"]
    ?? throw new InvalidOperationException("GREENER:HOST is not configured");
builder.Services.AddScoped(sp => new HttpClient { BaseAddress = new Uri(apiBaseUrl) });

builder.Services.AddScoped<ApiClient>();
builder.Services.AddScoped<AuthService>();
builder.Services.AddScoped<TestcaseService>();
builder.Services.AddScoped<SessionService>();
builder.Services.AddScoped<LabelService>();
builder.Services.AddScoped<ApiKeyService>();
builder.Services.AddScoped<GroupService>();

builder.Services.AddScoped<UserSessionService>();
builder.Services.AddScoped<ExceptionService>();
builder.Services.AddScoped<AppAuthenticationStateProvider>();
builder.Services.AddScoped<AuthenticationStateProvider>(sp =>
    sp.GetRequiredService<AppAuthenticationStateProvider>()
);

builder.Services.AddFluentUIComponents();
builder.Services.AddAuthorizationCore();
builder.Services.AddCascadingAuthenticationState();
builder.Services.AddBlazoredLocalStorage();

await builder.Build().RunAsync();
