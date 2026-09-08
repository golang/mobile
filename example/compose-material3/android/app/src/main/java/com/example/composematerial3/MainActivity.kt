package com.example.composematerial3

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import com.example.composematerial3.ui.theme.ComposeMaterial3Theme
import hello.Hello

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            ComposeMaterial3Theme {
                // A surface container using the 'background' color from the theme
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    Greeting("Android and Gopher")
                }
            }
        }
    }
}

@Composable
fun Greeting(name: String) {
    Text(text = Hello.greetings(name))
}

@Preview(showBackground = true)
@Composable
fun DefaultPreview() {
    ComposeMaterial3Theme {
        Greeting("Android and Gopher")
    }
}
