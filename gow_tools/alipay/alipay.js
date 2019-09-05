var webdriver = require('selenium-webdriver');
const { Builder, By, Key, until, Capabilities } = require('selenium-webdriver');
var chromeCapabilities = webdriver.Capabilities.chrome();
let nextPort = 20009;
//setting chrome options to start the browser fully maximized
var chromeOptions = {
    'args': ["--no-sandbox", "--disable-gpu"]
};
console.log('nextPort = ' + nextPort)
chromeCapabilities.set('chromeOptions', chromeOptions);
var driver = new webdriver.Builder().withCapabilities(chromeCapabilities).build();
let account = "63-9213177796";
let pwd = "aa147258";
  driver.get('https://auth.alipay.com/login/index.htm');
try {
  driver.getPageSource().then(src=>{
        console.log(src);
  });

  driver.manage().getCookie("ALIPAYJSESSIONID").then(cookies=>{
      console.log("cookies:",cookies);
  });

    let login = driver.findElement(By.id("J-loginMethod-tabs"));
  let login_ctx = login.findElements(By.css('li'));
  webdriver.promise.filter(login_ctx, function(li) {
    return li.getText().then(function(text) {
        console.log('Text value is: ' + text);
        return text === "账密登录";
    });
}).then(function (filteredSpans) {
    filteredSpans[0].click();
    var input_user = driver.findElement(By.id('J-input-user'));
    //console.log(input_user);
    input_user.sendKeys(account).then(sendkey => {
      input_user.getAttribute("value").then(user_value => {
        console.log("value:",user_value);
        var input_pwd =  driver.findElement(By.name('password_input'));
        for(let i in pwd){
          input_pwd.sendKeys(pwd.charAt(i));
        }
        var input_pwd2 =  driver.findElement(By.name('password_rsainput'));
        for(let i in pwd){
          input_pwd2.sendKeys(pwd.charAt(i));
        }
        input_pwd.getAttribute("value").then(pwd_value => {
        console.log("pwd value:",pwd_value);
        var form = driver.findElement(By.id('J-login-btn'));
        form.click().then(function(){
          console.log("form submit");
          driver.manage().getCookie("ALIPAYJSESSIONID").then(cookies=>{
              console.log("cookies:",cookies);
          });
          driver.getPageSource().then(src=>{
                console.log(src);
          });
        }).catch(err => {
          console.log("pwd 2 error:",err.message);
        });
        }).catch(err => {
          console.log("pwd 2 error:",err.message);
        });
        //input_pwd.sendKeys(pwd.charAt(0));

        /*input_pwd.sendKeys(pwd).then(function(){

        }).catch(function(err){
          console.log("pwd error:",err.message);
        });*/
      })
      .catch(err => {
        console.log("error:",err.message);
      });
    })
    .catch(err => {
      console.log("error:",err.message);
    });
});

  //input_user.sendKeys('"63-9213177796');
} catch (err) {
  console.log("Err:",err);
} finally {
  //driver.quit();
}
