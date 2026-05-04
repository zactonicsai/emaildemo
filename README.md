"# emaildemo" 

To connect your Gmail to Gemini, you need to enable the **Google Workspace extension**. This allows Gemini to find, summarize, and answer questions using your emails.

### **How to Enable Gmail Integration**

1.  **Open Gemini:** Go to the [Gemini web app](https://gemini.google.com/) or open the Gemini app on your phone.
2.  **Go to Settings:** Click on the **Settings** icon (gear icon) at the bottom left (web) or tap your profile picture and select **Extensions** (mobile).
3.  **Find Google Workspace:** Locate the **Google Workspace** toggle.
4.  **Turn it On:** Switch the toggle to **On**. You will see a prompt explaining what data Gemini will access; select **Connect**.
5.  **Verify Permissions:** Ensure that **Gmail** is checked within the Google Workspace sub-menu.

---

### **How to Use It**
Once connected, you can use the `@Gmail` command in your prompt or simply ask questions naturally:

* *"@Gmail summarize my most recent emails from the last 24 hours."*
* *"Find the tracking number for my latest order in my email."*
* *"What time is my flight next Tuesday according to my emails?"*

> **Note:** If you are using a Google Workspace account (like for work or school), your administrator may need to enable "Google Workspace Extensions" in the Admin Console before you can turn it on yourself.

### **Add Your Email as a Test User**

1. Go back to the [Google Cloud Console](https://console.cloud.google.com/).
2. Make sure your project (e.g., *your APP*) is selected in the top dropdown.
3. On the left sidebar, navigate to **APIs & Services** > **OAuth consent screen**.
4. Scroll down to the section labeled **Test users**.
5. Click the **+ Add Users** button.
6. Type in the exact email address you are using to authenticate (which looks like **your-mail@gmail.com** based on the previous error).
7. Click **Save**.

### **Try Again**
Once you have added your email to that list, go back to your terminal and run the app again:

```bash
go run main.go ./my_emails
```

When Chrome opens this time, Google will still give you a scary-looking warning saying "Google hasn’t verified this app." 

