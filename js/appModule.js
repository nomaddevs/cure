var app = angular.module('cure', ['ngRoute', 'ngCookies']);

app.config(function($routeProvider, $locationProvider) {
        $routeProvider
        .when("/", {
            templateUrl: "html/load.html"
        });

        $routeProvider.otherwise({
		    redirectTo: '/'
		});

        $locationProvider.html5Mode(true);
        $locationProvider.hashPrefix('');
});

app.controller("cureController", ['$scope', '$http', '$cookies', '$location', function($scope, $http, $cookies, $location) {
	$scope.Service = {};

	$scope.Service.Loading = true;
	$scope.Service.LoadSucceeded = false;

	$http.get("http://www.google.com")
	.then(function(response) {
		$scope.Service.Loading = false;
		$scope.Service.LoadSucceeded = true;
	}, function(response) {
		$scope.Service.Loading = false;
		$scope.Service.LoadSucceeded = false; 
	})

	$scope.ChangeView = function(view) {
	 	$location.path(view);
	};
}]);

app.directive('cureHeader', function() {
	return {
		templateUrl: './html/header.html',
	};
})
.directive('cureFooter', function() {
	return {
		templateUrl: './html/footer.html',
	};
})
.directive('cureContent', function() {
	return {
		templateUrl: function(elem, attr) {
			return "./html/" + attr.appPage + ".html";
		}
	};
});
