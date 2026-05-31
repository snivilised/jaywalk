package movies

var (
	brailleWaveFrames = []string{
		"⠁⠂⠄⡀", "⠂⠄⡀⢀", "⠄⡀⢀⠠", "⡀⢀⠠⠐",
		"⢀⠠⠐⠈", "⠠⠐⠈⠁", "⠐⠈⠁⠂", "⠈⠁⠂⠄",
	}

	dnaFrames = []string{
		"⠋⠉⠙⠚", "⠉⠙⠚⠒", "⠙⠚⠒⠂", "⠚⠒⠂⠂",
		"⠒⠂⠂⠒", "⠂⠂⠒⠲", "⠂⠒⠲⠴", "⠒⠲⠴⠤",
		"⠲⠴⠤⠄", "⠴⠤⠄⠋", "⠤⠄⠋⠉", "⠄⠋⠉⠙",
	}

	scanFrames = []string{
		"⠀⠀⠀⠀", "⡇⠀⠀⠀", "⣿⠀⠀⠀", "⢸⡇⠀⠀",
		"⠀⣿⠀⠀", "⠀⢸⡇⠀", "⠀⠀⣿⠀", "⠀⠀⢸⡇",
		"⠀⠀⠀⣿", "⠀⠀⠀⢸",
	}

	rainFrames = []string{
		"⢁⠂⠔⠈", "⠂⠌⡠⠐", "⠄⡐⢀⠡", "⡈⠠⠀⢂",
		"⠐⢀⠁⠄", "⠠⠁⠊⡀", "⢁⠂⠔⠈", "⠂⠌⡠⠐",
		"⠄⡐⢀⠡", "⡈⠠⠀⢂", "⠐⢀⠁⠄", "⠠⠁⠊⡀",
	}

	scanLineFrames = []string{
		"⠉⠉⠉", "⠓⠓⠓", "⠦⠦⠦", "⣄⣄⣄", "⠦⠦⠦", "⠓⠓⠓",
	}

	braillePulseFrames = []string{
		"⠀⠶⠀", "⠰⣿⠆", "⢾⣉⡷", "⣏⠀⣹", "⡁⠀⢈",
	}

	snakeFrames = []string{
		"⣁⡀", "⣉⠀", "⡉⠁", "⠉⠉", "⠈⠙", "⠀⠛",
		"⠐⠚", "⠒⠒", "⠖⠂", "⠶⠀", "⠦⠄", "⠤⠤",
		"⠠⢤", "⠀⣤", "⢀⣠", "⣀⣀",
	}

	sparkleFrames = []string{
		"⡡⠊⢔⠡", "⠊⡰⡡⡘", "⢔⢅⠈⢢",
		"⡁⢂⠆⡍", "⢔⠨⢑⢐", "⠨⡑⡠⠊",
	}

	cascadeFrames = []string{
		"⠀⠀⠀⠀", "⠀⠀⠀⠀", "⠁⠀⠀⠀", "⠋⠀⠀⠀",
		"⠞⠁⠀⠀", "⡴⠋⠀⠀", "⣠⠞⠁⠀", "⢀⡴⠋⠀",
		"⠀⣠⠞⠁", "⠀⢀⡴⠋", "⠀⠀⣠⠞", "⠀⠀⢀⡴",
		"⠀⠀⠀⣠", "⠀⠀⠀⢀",
	}

	columnsFrames = []string{
		"⡀⠀⠀", "⡄⠀⠀", "⡆⠀⠀", "⡇⠀⠀",
		"⣇⠀⠀", "⣧⠀⠀", "⣷⠀⠀", "⣿⠀⠀",
		"⣿⡀⠀", "⣿⡄⠀", "⣿⡆⠀", "⣿⡇⠀",
		"⣿⣇⠀", "⣿⣧⠀", "⣿⣷⠀", "⣿⣿⠀",
		"⣿⣿⡀", "⣿⣿⡄", "⣿⣿⡆", "⣿⣿⡇",
		"⣿⣿⣇", "⣿⣿⣧", "⣿⣿⣷", "⣿⣿⣿",
		"⣿⣿⣿", "⠀⠀⠀",
	}

	orbitFrames = []string{
		"⠃", "⠉", "⠘", "⠰", "⢠", "⣀", "⡄", "⠆",
	}

	breatheFrames = []string{
		"⠀", "⠂", "⠌", "⡑", "⢕", "⢝", "⣫", "⣟",
		"⣿", "⣟", "⣫", "⢝", "⢕", "⡑", "⠌", "⠂", "⠀",
	}

	waveRowsFrames = []string{
		"⠖⠉⠉⠑", "⡠⠖⠉⠉", "⣠⡠⠖⠉", "⣄⣠⡠⠖",
		"⠢⣄⣠⡠", "⠙⠢⣄⣠", "⠉⠙⠢⣄", "⠊⠉⠙⠢",
		"⠜⠊⠉⠙", "⡤⠜⠊⠉", "⣀⡤⠜⠊", "⢤⣀⡤⠜",
		"⠣⢤⣀⡤", "⠑⠣⢤⣀", "⠉⠑⠣⢤", "⠋⠉⠑⠣",
	}

	checkerboardFrames = []string{
		"⢕⢕⢕", "⡪⡪⡪", "⢊⠔⡡", "⡡⢊⠔",
	}

	helixFrames = []string{
		"⢌⣉⢎⣉", "⣉⡱⣉⡱", "⣉⢎⣉⢎", "⡱⣉⡱⣉",
		"⢎⣉⢎⣉", "⣉⡱⣉⡱", "⣉⢎⣉⢎", "⡱⣉⡱⣉",
		"⢎⣉⢎⣉", "⣉⡱⣉⡱", "⣉⢎⣉⢎", "⡱⣉⡱⣉",
		"⢎⣉⢎⣉", "⣉⡱⣉⡱", "⣉⢎⣉⢎", "⡱⣉⡱⣉",
	}

	fillSweepFrames = []string{
		"⣀⣀", "⣤⣤", "⣶⣶", "⣿⣿",
		"⣿⣿", "⣿⣿", "⣶⣶", "⣤⣤",
		"⣀⣀", "⠀⠀", "⠀⠀",
	}

	diagSwipeFrames = []string{
		"⠁⠀", "⠋⠀", "⠟⠁", "⡿⠋",
		"⣿⠟", "⣿⡿", "⣿⣿", "⣿⣿",
		"⣾⣿", "⣴⣿", "⣠⣾", "⢀⣴",
		"⠀⣠", "⠀⢀", "⠀⠀", "⠀⠀",
	}
)

var (
	brailleWaveSpinner  = frameArraySpinner(brailleWaveFrames)
	dnaSpinner          = frameArraySpinner(dnaFrames)
	scanSpinner         = frameArraySpinner(scanFrames)
	rainSpinner         = frameArraySpinner(rainFrames)
	scanLineSpinner     = frameArraySpinner(scanLineFrames)
	braillePulseSpinner = frameArraySpinner(braillePulseFrames)
	snakeSpinner        = frameArraySpinner(snakeFrames)
	sparkleSpinner      = frameArraySpinner(sparkleFrames)
	cascadeSpinner      = frameArraySpinner(cascadeFrames)
	columnsSpinner      = frameArraySpinner(columnsFrames)
	orbitSpinner        = frameArraySpinner(orbitFrames)
	breatheSpinner      = frameArraySpinner(breatheFrames)
	waveRowsSpinner     = frameArraySpinner(waveRowsFrames)
	checkerboardSpinner = frameArraySpinner(checkerboardFrames)
	helixSpinner        = frameArraySpinner(helixFrames)
	fillSweepSpinner    = frameArraySpinner(fillSweepFrames)
	diagSwipeSpinner    = frameArraySpinner(diagSwipeFrames)
)
