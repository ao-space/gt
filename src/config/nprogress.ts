import NProgress from "nprogress";
import "nprogress/nprogress.css";

NProgress.configure({
  easing: "ease", // Animation easing style
  speed: 500, // Speed at which the progress bar increases
  showSpinner: true, // Determines whether to display the loading spinner icon
  trickleSpeed: 200, // Interval for auto-incrementing the progress
  minimum: 0.3 // Minimum percentage when initialized
});

export default NProgress;
